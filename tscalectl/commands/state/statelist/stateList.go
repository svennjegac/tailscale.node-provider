package statelist

import (
	"fmt"
	"sort"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/svennjegac/tailscale.node-provider/internal/state"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show state",
	Long:  "Show currently deployed VPN nodes in AWS.",
	Args:  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) (runErr error) {
		defer func() {
			if r := recover(); r != nil {
				switch x := r.(type) {
				case error:
					runErr = errors.New(x.Error() + "\n")
					return
				default:
					runErr = errors.Errorf("unknown recovered type; val=%+v, type=%T", x, x)
				}
			}
		}()

		s := state.GetState()
		fmt.Printf("You have %d AWS nodes deployed.\n\n", len(s.Nodes))

		nodes := make([]string, 0, len(s.Nodes))
		for _, node := range s.Nodes {
			nodes = append(nodes, node.TscalectlName)
		}

		sort.Strings(nodes)

		for i, node := range nodes {
			fmt.Printf("%d - %s\n", i, node)
		}

		return nil
	},
}
