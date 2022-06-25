package creds

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pkg/errors"

	"github.com/svennjegac/tailscale.node-provider/internal/fileutil"
	"github.com/svennjegac/tailscale.node-provider/internal/tscos"
)

type Creds struct {
	AwsAccessKeyID     string `json:"aws_access_key_id"`
	AwsSecretAccessKey string `json:"aws_secret_access_key"`
	TailscaleAuthKey   string `json:"tailscale_auth_key"`
}

func Get() Creds {
	fileutil.MkdirAllFromFile(tscos.CredsFile())

	unlock := fileutil.Lock(tscos.CredsFile())
	defer unlock()

	if _, err := os.Stat(tscos.CredsFile()); os.IsNotExist(err) {
		return userInputCreds()

	} else if err != nil {
		panic(err)
	} else {
		b := fileutil.ReadFile(tscos.CredsFile())
		if len(b) == 0 {
			return userInputCreds()
		}

		var creds Creds
		err = json.Unmarshal(b, &creds)
		if err != nil {
			panic(errors.Wrap(err, "json unmarshal creds file"))
		}

		return creds
	}
}

func userInputCreds() Creds {
	fmt.Println("Your credentials are not set up, please provide them")

	fmt.Printf("aws_access_key_id: ")
	var awsAccessKeyID string
	_, err := fmt.Fscanln(os.Stdin, &awsAccessKeyID)
	if err != nil {
		panic(errors.Wrap(err, "error scanning aws access key id"))
	}

	fmt.Printf("aws_secret_access_key: ")
	var awsSecretAccessKey string
	_, err = fmt.Fscanln(os.Stdin, &awsSecretAccessKey)
	if err != nil {
		panic(errors.Wrap(err, "error scanning aws secret access key"))
	}

	fmt.Printf("tailscale_auth_key: ")
	var tailscaleAuthKey string
	_, err = fmt.Fscanln(os.Stdin, &tailscaleAuthKey)
	if err != nil {
		panic(errors.Wrap(err, "error scanning tailscale auth key"))
	}

	fmt.Println()

	creds := Creds{
		AwsAccessKeyID:     awsAccessKeyID,
		AwsSecretAccessKey: awsSecretAccessKey,
		TailscaleAuthKey:   tailscaleAuthKey,
	}

	credsBytes, err := json.Marshal(creds)
	if err != nil {
		panic(errors.Wrap(err, "json marshal creds"))
	}
	fileutil.WriteFile(tscos.CredsFile(), credsBytes)

	return creds
}
