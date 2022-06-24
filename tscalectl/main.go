package main

import (
	"github.com/svennjegac/tailscale.node-provider/tscalectl/commands"
)

func main() {
	commands.RootCmd.Execute()
}
