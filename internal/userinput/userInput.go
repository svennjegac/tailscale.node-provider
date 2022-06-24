package userinput

import (
	"fmt"
	"os"

	"github.com/pkg/errors"

	"github.com/svennjegac/tailscale.node-provider/internal/aws/ec2cli"
)

func Region(interactiveFlag bool, regionFlag string) string {
	regions := ec2cli.Regions()

	if len(regionFlag) > 0 {
		for _, r := range regions {
			if r == regionFlag {
				return regionFlag
			}
		}
		panic(errors.Errorf("user input region, provided invalid region as flag, use one of the allowed regions; "+
			"region-flag=%s, allowed-regions=%+v", regionFlag, regions))
	}

	if !interactiveFlag {
		panic(errors.New("user input region, please specify region or use interactive flag"))
	}

	fmt.Println("Allowed regions:")
	for i, r := range regions {
		fmt.Printf("%2d - %s\n", i, r)
	}

	fmt.Println()
	fmt.Println("Please enter number representing the region.")
	fmt.Printf("region: ")
	var region int
	_, err := fmt.Fscanln(os.Stdin, &region)
	if err != nil {
		panic(errors.Wrap(err, "user input region, failed to read user input"))
	}

	if region < 0 || region >= len(regions) {
		panic(errors.New("please enter one of the allowed region numbers"))
	}

	fmt.Println()
	return regions[region]
}

func InstanceType(interactiveFlag bool, instanceTypeFlag string, region string) string {
	instanceTypes := ec2cli.InstanceTypesPerRegion(region)

	if len(instanceTypeFlag) > 0 {
		for _, it := range instanceTypes {
			if it == instanceTypeFlag {
				return instanceTypeFlag
			}
		}
		panic(errors.Errorf("user input instance type, provided invalid instance type as flag, use one of the allowed instance types; "+
			"instance-type-flag=%s, allowed-instance-types=%+v", instanceTypeFlag, instanceTypes))
	}

	if !interactiveFlag {
		panic(errors.New("user input instance type, please specify instance type or use interactive flag"))
	}

	fmt.Println("Allowed instance types:")
	for i, it := range instanceTypes {
		fmt.Printf("%3d - %s\n", i, it)
	}

	fmt.Println()
	fmt.Println("Please enter number representing the instance type.")
	fmt.Printf("instance_type: ")
	var instanceType int
	_, err := fmt.Fscanln(os.Stdin, &instanceType)
	if err != nil {
		panic(errors.Wrap(err, "user input instance type, failed to read user input"))
	}

	if instanceType < 0 || instanceType >= len(instanceTypes) {
		panic(errors.New("please enter one of the allowed instance type numbers"))
	}

	fmt.Println()

	return instanceTypes[instanceType]
}

func AMI(interactiveFlag bool, amiFlag string, region string) string {
	if len(amiFlag) > 0 {
		return amiFlag
	}

	if !interactiveFlag {
		panic(errors.New("user input AMI, please specify AMI or use interactive flag"))
	}

	amis := ec2cli.AMIsPerRegion(region)

	fmt.Println("Allowed AMIs:")
	for i, ami := range amis {
		fmt.Printf("%1d - %s\n", i, ami)
	}
	fmt.Println("(There is too many AMIs to list them all, if you want to use other AMI, specify it through the AMI flag)")

	fmt.Println()
	fmt.Println("Please enter number representing the AMI.")
	fmt.Printf("ami: ")
	var ami int
	_, err := fmt.Fscanln(os.Stdin, &ami)
	if err != nil {
		panic(errors.Wrap(err, "user input AMI, failed to read user input"))
	}

	if ami < 0 || ami >= len(amis) {
		panic(errors.New("please enter one of the allowed AMI numbers"))
	}

	fmt.Println()

	return amis[ami]
}
