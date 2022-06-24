package tscos

import (
	"os"

	"github.com/pkg/errors"
)

var homeDir string

func init() {
	hd, err := os.UserHomeDir()
	if err != nil {
		panic(errors.Wrap(err, "os user home dir"))
	}
	homeDir = hd
}

func HomeDir() string {
	return homeDir
}

func TscalectlDir() string {
	return HomeDir() + "/.tscalectl"
}

func AwsKeyPairsDir() string {
	return TscalectlDir() + "/awskeypairs"
}

func CredsFile() string {
	return TscalectlDir() + "/credentials.json"
}

func StateFile() string {
	return TscalectlDir() + "/state.json"
}

func KnownHostsFile() string {
	return TscalectlDir() + "/known_hosts"
}
