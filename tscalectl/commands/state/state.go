package state

import (
	"github.com/spf13/cobra"

	"github.com/svennjegac/tailscale.node-provider/tscalectl/commands/state/statedump"
	"github.com/svennjegac/tailscale.node-provider/tscalectl/commands/state/statelist"
)

var StateCmd = &cobra.Command{
	Use:   "state",
	Short: "Manage CLI state",
	Long:  "Manage CLI state.",
	Args:  cobra.ExactArgs(0),
}

func init() {
	StateCmd.AddCommand(statedump.DumpCmd)
	StateCmd.AddCommand(statelist.ListCmd)
}
