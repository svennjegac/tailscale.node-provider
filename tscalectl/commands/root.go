package commands

import (
	"github.com/spf13/cobra"

	"github.com/svennjegac/tailscale.node-provider/tscalectl/commands/down"
	"github.com/svennjegac/tailscale.node-provider/tscalectl/commands/state"
	"github.com/svennjegac/tailscale.node-provider/tscalectl/commands/up"
)

var RootCmd = &cobra.Command{
	Use:     "tscalectl",
	Version: "v0.0.0",
}

func init() {
	RootCmd.AddCommand(up.UpCmd)
	RootCmd.AddCommand(down.DownCmd)
	RootCmd.AddCommand(state.StateCmd)
}
