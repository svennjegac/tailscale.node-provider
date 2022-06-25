package statedump

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/svennjegac/tailscale.node-provider/internal/state"
	"github.com/svennjegac/tailscale.node-provider/internal/trycatch"
)

var DumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump state JSON",
	Long:  "Dump state JSON and its internal details",
	Args:  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) (runErr error) {
		defer trycatch.ToError(&runErr)

		s := state.GetState()

		b, err := json.MarshalIndent(s, "", "  ")
		if err != nil {
			panic(errors.Wrap(err, "state dump, json marshal"))
		}

		fmt.Println(string(b))

		return nil
	},
}
