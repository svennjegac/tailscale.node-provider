package creds

import (
	"github.com/spf13/cobra"

	credsdelete "github.com/svennjegac/tailscale.node-provider/tscalectl/commands/creds/delete"
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
