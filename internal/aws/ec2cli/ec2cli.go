package ec2cli

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"

	"github.com/svennjegac/tailscale.node-provider/internal/creds"
)

var ec2Client *ec2.Client

func init() {
	crd := creds.Get()
	ec2Client = ec2.New(ec2.Options{
		Credentials: aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     crd.AwsAccessKeyID,
				SecretAccessKey: crd.AwsSecretAccessKey,
			}, nil
		}),
	})
}

func Regions() []string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	res, err := ec2Client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{
		AllRegions:  nil,
		DryRun:      nil,
		Filters:     nil,
		RegionNames: nil,
	}, func(options *ec2.Options) {
		options.Region = "eu-central-1"
	})
	if err != nil {
		panic(errors.Wrap(err, "ec2cli regions"))
	}

	regionNames := make([]string, 0, len(res.Regions))
	for _, r := range res.Regions {
		regionNames = append(regionNames, *r.RegionName)
	}

	sort.Strings(regionNames)

	return regionNames
}

func InstanceTypesPerRegion(region string) []string {
	instanceTypes := make([]string, 0, 20)
	var nextToken *string
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

		res, err := ec2Client.DescribeInstanceTypeOfferings(ctx, &ec2.DescribeInstanceTypeOfferingsInput{
			DryRun: nil,
			Filters: []types.Filter{
				{
					Name:   aws.String("location"),
					Values: []string{region},
				},
			},
			NextToken: nextToken,
		}, func(options *ec2.Options) {
			options.Region = region
		})
		cancel()
		if err != nil {
			panic(errors.Wrap(err, "ec2cli instance types per region"))
		}

		nextToken = res.NextToken

		for _, offering := range res.InstanceTypeOfferings {
			instanceTypes = append(instanceTypes, string(offering.InstanceType))
		}

		if nextToken == nil {
			break
		}
	}

	sort.Strings(instanceTypes)

	return instanceTypes
}

func AMIsPerRegion(region string) []string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	res, err := ec2Client.DescribeImages(ctx, &ec2.DescribeImagesInput{
		DryRun:          nil,
		ExecutableUsers: nil,
		Filters: []types.Filter{
			{
				Name:   aws.String("name"),
				Values: []string{"ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-20220609"},
			},
			{
				Name:   aws.String("owner-id"),
				Values: []string{"099720109477"},
			},
		},
		ImageIds:          nil,
		IncludeDeprecated: nil,
		Owners:            nil,
	}, func(options *ec2.Options) {
		options.Region = region
	})
	if err != nil {
		panic(errors.Wrap(err, "ec2cli AMIs per region"))
	}

	images := make([]string, 0, len(res.Images))
	for _, image := range res.Images {
		images = append(images, *image.ImageId)
	}

	sort.Strings(images)

	return images
}

func ImportKeyPair(region string, keyName string, pubKey ssh.PublicKey) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	_, err := ec2Client.ImportKeyPair(ctx, &ec2.ImportKeyPairInput{
		KeyName:           aws.String(keyName),
		PublicKeyMaterial: ssh.MarshalAuthorizedKey(pubKey),
		DryRun:            nil,
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeKeyPair,
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(keyName),
					},
					{
						Key:   aws.String("Description"),
						Value: aws.String("tailscalectl managed key pair"),
					},
				},
			},
		},
	}, func(options *ec2.Options) {
		options.Region = region
	})
	if err != nil {
		panic(errors.Wrap(err, "ec2cli import key pair"))
	}
}

func CreateSecurityGroup(region string, securityGroupName string) string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	secGrOut, err := ec2Client.CreateSecurityGroup(ctx, &ec2.CreateSecurityGroupInput{
		Description: aws.String("tailscalectl managed security group"),
		GroupName:   aws.String(securityGroupName),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeSecurityGroup,
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(securityGroupName),
					},
					{
						Key:   aws.String("Description"),
						Value: aws.String("tailscalectl managed security group"),
					},
				},
			},
		},
		VpcId: nil,
	}, func(options *ec2.Options) {
		options.Region = region
	})
	if err != nil {
		panic(errors.Wrap(err, "ec2cli create security group"))
	}

	_, err = ec2Client.AuthorizeSecurityGroupIngress(ctx, &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: secGrOut.GroupId,
		IpPermissions: []types.IpPermission{
			{
				FromPort:   aws.Int32(22),
				IpProtocol: aws.String("tcp"),
				IpRanges: []types.IpRange{
					{
						CidrIp:      aws.String("0.0.0.0/0"),
						Description: aws.String("Allow only SSH connections"),
					},
				},
				ToPort: aws.Int32(22),
			},
		},
	}, func(options *ec2.Options) {
		options.Region = region
	})
	if err != nil {
		panic(errors.Wrap(err, "ec2cli authorize security group ingress"))
	}

	return *secGrOut.GroupId
}

func RunInstance(region string, instanceType string, ami string, vpnNodeName string, securityGroupID string) string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	runInstOut, err := ec2Client.RunInstances(ctx, &ec2.RunInstancesInput{
		MaxCount:         aws.Int32(1),
		MinCount:         aws.Int32(1),
		ImageId:          aws.String(ami),
		InstanceType:     types.InstanceType(instanceType),
		KeyName:          aws.String(vpnNodeName),
		SecurityGroupIds: []string{securityGroupID},
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeInstance,
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(vpnNodeName),
					},
					{
						Key:   aws.String("Description"),
						Value: aws.String("tailscalectl managed ec2 instance"),
					},
				},
			},
		},
	}, func(options *ec2.Options) {
		options.Region = region
	})
	if err != nil {
		panic(errors.Wrap(err, "ec2cli run instances"))
	}

	return *runInstOut.Instances[0].InstanceId
}

func TerminateInstance(region string, vpnNodeName string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	descInstOut, err := ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:Name"),
				Values: []string{vpnNodeName},
			},
		},
	}, func(options *ec2.Options) {
		options.Region = region
	})
	if err != nil {
		panic(errors.Wrap(err, "ec2cli terminate instances, descirbe"))
	}

	if len(descInstOut.Reservations) == 0 {
		return
	}

	termInstOut, err := ec2Client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []string{*descInstOut.Reservations[0].Instances[0].InstanceId},
		DryRun:      nil,
	}, func(options *ec2.Options) {
		options.Region = region
	})
	if err != nil {
		panic(errors.Wrap(err, "ec2cli terminate instances, terminate"))
	}

	if len(termInstOut.TerminatingInstances) != 1 {
		panic(errors.Errorf("terminate instance, terminating != 1 instance; num=%d", len(termInstOut.TerminatingInstances)))
	}
}

func DeleteSecurityGroup(region string, vpnNodeName string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, err := ec2Client.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
		GroupName: aws.String(vpnNodeName),
	}, func(options *ec2.Options) {
		options.Region = region
	})
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "does not exist in default VPC") {
			return
		}
		panic(errors.Wrap(err, "ec2cli delete security group"))
	}
}

func DeleteKeyPair(region string, vpnNodeName string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, err := ec2Client.DeleteKeyPair(ctx, &ec2.DeleteKeyPairInput{
		KeyName: aws.String(vpnNodeName),
	}, func(options *ec2.Options) {
		options.Region = region
	})
	if err != nil {
		panic(errors.Wrap(err, "ec2cli delete key pair"))
	}
}

func WaitForInstanceToInitialize(region string, ec2InstanceID string) {
	fmt.Println("Waiting for EC2 instance to boot")

	startTime := time.Now()
	for {
		time.Sleep(time.Second * 5)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

		statusOut, err := ec2Client.DescribeInstanceStatus(ctx, &ec2.DescribeInstanceStatusInput{
			InstanceIds: []string{ec2InstanceID},
		}, func(options *ec2.Options) {
			options.Region = region
		})
		cancel()
		if err != nil {
			panic(errors.Wrap(err, "ec2cli describe instance status"))
		}

		if len(statusOut.InstanceStatuses) < 1 {
			fmt.Println("No instance info yet, continuing to wait...", time.Since(startTime))
			continue
		}

		if statusOut.InstanceStatuses[0].InstanceStatus.Status == "initializing" && statusOut.InstanceStatuses[0].InstanceStatus.Details[0].Status == "initializing" {
			fmt.Println("Instance currently initializing, continuing to wait...", time.Since(startTime))
			continue
		} else if statusOut.InstanceStatuses[0].InstanceStatus.Status == "ok" && statusOut.InstanceStatuses[0].InstanceStatus.Details[0].Status == "passed" {
			fmt.Println("Instance ready", time.Since(startTime))
			break
		} else {
			fmt.Println("Instance not initialized properly", time.Since(startTime))
			panic(errors.New("ec2cli instance not initialized properly"))
		}
	}
}

func WaitForInstanceToTerminate(region string, vpnNodeName string) {
	startTime := time.Now()
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

		descInstOut, err := ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("tag:Name"),
					Values: []string{vpnNodeName},
				},
			},
		}, func(options *ec2.Options) {
			options.Region = region
		})
		cancel()
		if err != nil {
			panic(errors.Wrap(err, "ec2cli wait for instances to terminate, describe"))
		}

		if len(descInstOut.Reservations) == 0 {
			return
		}

		if len(descInstOut.Reservations[0].Instances) != 1 {
			panic(errors.Errorf("ec2cli wait for instances to terminate, wrong num of instances with requested vpn node name; num=%d", len(descInstOut.Reservations[0].Instances)))
		}

		if descInstOut.Reservations[0].Instances[0].State.Name != types.InstanceStateNameTerminated {
			fmt.Printf("Instance state: %s, waiting to terminate... %s\n", descInstOut.Reservations[0].Instances[0].State.Name, time.Since(startTime))
			time.Sleep(time.Second * 5)
		} else {
			fmt.Printf("Instance state: %s, %s\n", descInstOut.Reservations[0].Instances[0].State.Name, time.Since(startTime))
			return
		}
	}
}

func DescribeInstance(region string, ec2InstanceID string) string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	descOut, err := ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{ec2InstanceID},
	}, func(options *ec2.Options) {
		options.Region = region
	})
	if err != nil {
		panic(errors.Wrap(err, "ec2cli describe instance"))
	}

	return *descOut.Reservations[0].Instances[0].PublicIpAddress
}
