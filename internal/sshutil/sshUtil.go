package sshutil

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"

	"github.com/svennjegac/tailscale.node-provider/internal/fileutil"
	"github.com/svennjegac/tailscale.node-provider/internal/tscos"
)

func StartTailscale(privateKey *rsa.PrivateKey, ec2PublicIP string, tailscaleAuthKey string, hostname string, exitNode bool) {
	// ssh config
	hostKeyCallback, err := knownhosts.New(tscos.KnownHostsFile())
	if err != nil {
		panic(errors.Wrap(err, "start tailscale, host key callback"))
	}

	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		panic(errors.Wrap(err, "start tailscale, new signer from key"))
	}

	config := &ssh.ClientConfig{
		User: "ubuntu",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: hostKeyCallback,
		Timeout:         time.Second * 10,
	}

	// connect to ssh server
	client, err := ssh.Dial("tcp", ec2PublicIP+":22", config)
	if err != nil {
		panic(errors.Wrap(err, "start tailscale, ssh dial"))
	}
	defer client.Close()

	advertiseExitNodeFlag := "--advertise-exit-node"
	if !exitNode {
		advertiseExitNodeFlag = ""
	}

	execSSH(client, "curl -fsSL https://tailscale.com/install.sh | sh")
	execSSH(client, "echo 'net.ipv4.ip_forward = 1' | sudo tee -a /etc/sysctl.conf")
	execSSH(client, "echo 'net.ipv6.conf.all.forwarding = 1' | sudo tee -a /etc/sysctl.conf")
	execSSH(client, "sudo sysctl -p /etc/sysctl.conf")
	execSSH(client, fmt.Sprintf("sudo tailscale up --auth-key %s --hostname %s %s", tailscaleAuthKey, hostname, advertiseExitNodeFlag))
}

func execSSH(client *ssh.Client, command string) {
	session, err := client.NewSession()
	if err != nil {
		panic(errors.Wrap(err, "ssh client new session"))
	}
	defer session.Close()

	var buff bytes.Buffer
	session.Stdout = &buff
	if err = session.Run(command); err != nil {
		panic(errors.Wrap(err, "ssh session run command; command:"+command))
	}
	fmt.Println(buff.String())
}

func CreateKeyPair(keyName string) (*rsa.PrivateKey, ssh.PublicKey) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(errors.Wrap(err, "generate private RSA key"))
	}

	fileutil.MkdirAll(tscos.AwsKeyPairsDir())

	unlockPem := fileutil.Lock(tscos.AwsKeyPairsDir() + "/" + keyName + ".pem")
	defer unlockPem()

	// generate and write private key as PEM
	privateKeyFile, err := os.Create(tscos.AwsKeyPairsDir() + "/" + keyName + ".pem")
	if err != nil {
		panic(errors.Wrap(err, "create RSA pem file"))
	}
	defer privateKeyFile.Close()

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	if err = pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
		panic(errors.Wrap(err, "encode RSA to pem file"))
	}
	privateKeyFile.Chmod(0400)

	// generate and write public key
	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		panic(errors.Wrap(err, "public ssh key from rsa public key"))
	}

	unlockPub := fileutil.Lock(tscos.AwsKeyPairsDir() + "/" + keyName + ".pub")
	defer unlockPub()

	fileutil.WriteFilePerm(tscos.AwsKeyPairsDir()+"/"+keyName+".pub", ssh.MarshalAuthorizedKey(pub), 0644)

	return privateKey, pub
}

func DeleteKeyPair(keyName string) {
	fileutil.MkdirAll(tscos.AwsKeyPairsDir())

	unlockPem := fileutil.Lock(tscos.AwsKeyPairsDir() + "/" + keyName + ".pem")
	defer unlockPem()

	if _, err := os.Stat(tscos.AwsKeyPairsDir() + "/" + keyName + ".pem"); err == nil {
		fileutil.Remove(tscos.AwsKeyPairsDir() + "/" + keyName + ".pem")
	}

	unlockPub := fileutil.Lock(tscos.AwsKeyPairsDir() + "/" + keyName + ".pub")
	defer unlockPub()

	if _, err := os.Stat(tscos.AwsKeyPairsDir() + "/" + keyName + ".pub"); err == nil {
		fileutil.Remove(tscos.AwsKeyPairsDir() + "/" + keyName + ".pub")
	}
}

func UpdateKnownHosts(privKey *rsa.PrivateKey, ec2InstancePublicIP string) {
	signer, err := ssh.NewSignerFromKey(privKey)
	if err != nil {
		panic(errors.Wrap(err, "update known hosts, new signer from key"))
	}

	config := &ssh.ClientConfig{
		User: "ubuntu",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: keyPrint,
		Timeout:         time.Second * 10,
	}

	// connect ot ssh server
	client, err := ssh.Dial("tcp", ec2InstancePublicIP+":22", config)
	if err != nil {
		panic(errors.Wrap(err, "update known hosts, dial ssh"))
	}
	defer client.Close()
}

func keyPrint(dialAddr string, addr net.Addr, key ssh.PublicKey) error {
	fileutil.MkdirAllFromFile(tscos.KnownHostsFile())

	unlock := fileutil.Lock(tscos.KnownHostsFile())
	defer unlock()

	f, err := os.OpenFile(tscos.KnownHostsFile(), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return errors.Wrap(err, "open known hosts file")
	}
	defer f.Close()

	// if known hosts didn't exist before, fileutil created it with wrong perm
	err = f.Chmod(0600)
	if err != nil {
		panic(errors.Wrap(err, "create known hosts, chmod"))
	}

	if _, err = f.WriteString(fmt.Sprintf("%s %s %s\n", strings.Split(dialAddr, ":")[0], key.Type(), base64.StdEncoding.EncodeToString(key.Marshal()))); err != nil {
		return errors.Wrap(err, "write to known hosts file")
	}

	return nil
}
