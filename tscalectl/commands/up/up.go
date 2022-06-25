package up

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/svennjegac/tailscale.node-provider/internal/aws/ec2cli"
	"github.com/svennjegac/tailscale.node-provider/internal/creds"
	"github.com/svennjegac/tailscale.node-provider/internal/sshutil"
	"github.com/svennjegac/tailscale.node-provider/internal/state"
	"github.com/svennjegac/tailscale.node-provider/internal/trycatch"
	"github.com/svennjegac/tailscale.node-provider/internal/userinput"
)

var interactiveFlag bool
var exitNodeFlag bool
var regionFlag string
var instanceTypeFlag string
var amiFlag string

var UpCmd = &cobra.Command{
	Use:   "up",
	Short: "Create new tailscale node",
	Long:  "Create new tailscale node with automatic authentication. There is no need for web approval.",
	// Args:  cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) (runErr error) {
		defer trycatch.ToError(&runErr)

		// user interaction
		region := userinput.Region(interactiveFlag, regionFlag)
		instanceType := userinput.InstanceType(interactiveFlag, instanceTypeFlag, region)
		ami := userinput.AMI(interactiveFlag, amiFlag, region)

		// CLI internal state
		vpnNode := state.AddNewNode(region, instanceType, ami)

		fmt.Printf("Starting VPN node provisioning (%s, %s, %s)\n", region, instanceType, ami)

		privK, pubK := sshutil.CreateKeyPair(vpnNode.TscalectlName)

		// AWS provisioning
		fmt.Println("Importing EC2 key pair")
		ec2cli.ImportKeyPair(region, vpnNode.TscalectlName, pubK)
		fmt.Println("Creating EC2 security group")
		securityGroupID := ec2cli.CreateSecurityGroup(region, vpnNode.TscalectlName)
		fmt.Println("Creating EC2 instance")
		ec2InstanceID := ec2cli.RunInstance(region, instanceType, ami, vpnNode.TscalectlName, securityGroupID)
		fmt.Println("Waiting for EC2 instance to boot")
		ec2cli.WaitForInstanceToInitialize(region, ec2InstanceID)
		ec2InstancePublicIP := ec2cli.DescribeInstance(region, vpnNode.TscalectlName)

		// starting tailscale on provisioned node
		crd := creds.Get()
		fmt.Println("Updating SSH known hosts")
		sshutil.UpdateKnownHosts(privK, ec2InstancePublicIP)
		fmt.Println("Starting tailscale")
		sshutil.StartTailscale(privK, ec2InstancePublicIP, crd.TailscaleAuthKey, vpnNode.TscalectlName, exitNodeFlag)

		fmt.Println("VPN node ready for use")

		return nil
	},
}

func init() {
	UpCmd.Flags().BoolVarP(&interactiveFlag, "interactive", "i", false, "CLI will help you to choose params which were not provided as flags (region, instance type, ami)")
	UpCmd.Flags().BoolVarP(&exitNodeFlag, "exit-node", "e", false, "Advertise VPN node as tailscale exit node")
	UpCmd.Flags().StringVarP(&regionFlag, "region", "r", "", "Region in which VPN node should be created (AWS region, e.g. eu-west-1)")
	UpCmd.Flags().StringVarP(&instanceTypeFlag, "instance-type", "t", "", "VPN node instance type (AWS instance type, e.g. t2.micro)")
	UpCmd.Flags().StringVarP(&amiFlag, "ami", "a", "", "VPN node ami (AWS amazon machine image (OS))")
}
