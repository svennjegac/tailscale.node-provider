package down

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/svennjegac/tailscale.node-provider/internal/aws/ec2cli"
	"github.com/svennjegac/tailscale.node-provider/internal/sshutil"
	"github.com/svennjegac/tailscale.node-provider/internal/state"
	"github.com/svennjegac/tailscale.node-provider/internal/trycatch"
)

var DownCmd = &cobra.Command{
	Use:   "down [nodeID string]",
	Short: "Terminate tailscale node and remove associated resources",
	Long:  "Terminate tailscale node and remove associated resources. (Remove it from state file, remove SSH keys, delete AWS instance, security group and key pairs)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (runErr error) {
		defer trycatch.ToError(&runErr)

		tscalectlID, err := strconv.Atoi(args[0])
		if err != nil {
			panic(errors.Wrap(err, "provide integer ID for node ID (first 3 numbers of your node name)"))
		}

		node := state.GetNode(tscalectlID)

		ec2cli.TerminateInstance(node.Region, node.TscalectlName)
		fmt.Println("Deleted EC2 instance")
		ec2cli.WaitForInstanceToTerminate(node.Region, node.TscalectlName)
		ec2cli.DeleteSecurityGroup(node.Region, node.TscalectlName)
		fmt.Println("Deleted EC2 security group")
		ec2cli.DeleteKeyPair(node.Region, node.TscalectlName)
		fmt.Println("Deleted EC2 key pair")

		sshutil.DeleteKeyPair(node.TscalectlName)
		fmt.Println("Deleted CLI local SSH keys")

		state.RemoveNode(tscalectlID)
		fmt.Println("Deleted node from CLI local state")

		return nil
	},
}
