package creds

import (
	"github.com/spf13/cobra"

	"github.com/svennjegac/tailscale.node-provider/tscalectl/commands/creds/credsdelete"
)

var CredsCmd = &cobra.Command{
	Use:   "creds",
	Short: "Manage CLI creds",
	Long:  "Manage CLI creds.",
	Args:  cobra.ExactArgs(0),
}

func init() {
	CredsCmd.AddCommand(credsdelete.DeleteCmd)
}
