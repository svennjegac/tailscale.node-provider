package credsdelete

import (
	"github.com/spf13/cobra"

	"github.com/svennjegac/tailscale.node-provider/internal/fileutil"
	"github.com/svennjegac/tailscale.node-provider/internal/trycatch"
	"github.com/svennjegac/tailscale.node-provider/internal/tscos"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete credentials",
	Long:  "Delete credentials. This will cause you to be prompted for new credentials on e.g. UP command. (If you dont create credentials before)",
	Args:  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) (runErr error) {
		defer trycatch.ToError(&runErr)

		fileutil.MkdirAllFromFile(tscos.CredsFile())
		fileutil.Remove(tscos.CredsFile())

		return nil
	},
}
