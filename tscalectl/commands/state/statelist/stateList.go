package statelist

import (
	"fmt"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/svennjegac/tailscale.node-provider/internal/state"
	"github.com/svennjegac/tailscale.node-provider/internal/trycatch"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show state",
	Long:  "Show currently deployed tailscale nodes in AWS.",
	Args:  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) (runErr error) {
		defer trycatch.ToError(&runErr)

		s := state.GetState()
		fmt.Printf("You have %d AWS nodes deployed.\n\n", len(s.Nodes))

		nodes := make([]*state.VPNNode, 0, len(s.Nodes))
		for _, node := range s.Nodes {
			nodes = append(nodes, node)
		}

		sort.Slice(nodes, func(i, j int) bool {
			if nodes[i].TscalectlID < nodes[j].TscalectlID {
				return true
			}
			return false
		})

		for i, node := range nodes {
			fmt.Printf("%d - %s, age: %s\n", i, node.TscalectlName, time.Since(node.CreatedAt))
		}

		return nil
	},
}
