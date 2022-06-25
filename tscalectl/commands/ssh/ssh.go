package ssh

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/svennjegac/tailscale.node-provider/internal/aws/ec2cli"
	"github.com/svennjegac/tailscale.node-provider/internal/state"
	"github.com/svennjegac/tailscale.node-provider/internal/trycatch"
	"github.com/svennjegac/tailscale.node-provider/internal/tscos"
)

var SSHCmd = &cobra.Command{
	Use:   "ssh [nodeID string]",
	Short: "Print SSH access to VPN node",
	Long:  "Command prints you c/p command which can be used for accessing VPN node. It will find appropriate SSH key and IP address.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (runErr error) {
		defer trycatch.ToError(&runErr)

		tscalectlID, err := strconv.Atoi(args[0])
		if err != nil {
			panic(errors.Wrap(err, "provide integer ID for node ID (first 3 numbers of your node name)"))
		}

		node := state.GetNode(tscalectlID)

		ec2InstancePublicIP := ec2cli.DescribeInstance(node.Region, node.TscalectlName)

		fmt.Printf("ssh -tt -i %s ubuntu@%s\n", tscos.AwsKeyPairsDir()+"/"+node.TscalectlName+".pem", ec2InstancePublicIP)

		return nil
	},
}
