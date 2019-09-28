package ssh

import (
	"bytes"
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg/k3pi"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"io/ioutil"
	"net"
	"os"
	"sync"
	"time"
)

// SSH settings to use.
type Settings struct {
	User, KeyPath, Port string
}

// Returns the expanded key path.
func (s *Settings) GetKeyPath() string {
	path, _ := homedir.Expand(s.KeyPath)
	return path
}

/* Loads the ssh public key.
   LoadPublicKey(Settings{KeyPath: "~/.ssh/id_rsa"})
*/
func LoadPublicKey(settings *Settings) (ssh.AuthMethod, func() error, error) {
	noopCloseFunc := func() error { return nil }
	keyPath := settings.GetKeyPath()

	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, noopCloseFunc, err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		if err.Error() != "ssh: cannot decode encrypted private keys" {
			return nil, noopCloseFunc, err
		}

		agent, close := sshAgent(keyPath + ".pub")
		if agent != nil {
			return agent, close, nil
		}

		defer close()

		fmt.Printf("Enter passphrase for '%s': ", keyPath)
		bytePassword, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()

		signer, err = ssh.ParsePrivateKeyWithPassphrase(key, bytePassword)
		if err != nil {
			return nil, noopCloseFunc, err
		}
	}

	return ssh.PublicKeys(signer), noopCloseFunc, nil
}

// Creates a new ssh client configuration.
func NewClientConfig(settings *Settings) (*ssh.ClientConfig, func() error, error) {
	authMethod, closeSSHAgent, err := LoadPublicKey(settings)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "unable to load the ssh key with path %q", settings.GetKeyPath())
	}

	return &ssh.ClientConfig{
		User: settings.User,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout: time.Second*3,
	}, closeSSHAgent, nil
}

func NewHypriotClientConfig() (*ssh.ClientConfig, func() error, error) {
	return &ssh.ClientConfig{
		User:              "pirate",
		Auth:              []ssh.AuthMethod{ ssh.Password("hypriot") },
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout: time.Second*3,
	}, func() error { return nil }, nil
}

type cmdRunner struct {
	writeToStdOut bool
	client *ssh.Client
}

func (s *cmdRunner) Close() error {
	return s.client.Close()
}

func (s *cmdRunner) Execute(command string) (*k3pi.Result, error) {
	sess, err := s.client.NewSession()
	if err != nil {
		return &k3pi.Result{}, err
	}

	defer sess.Close()

	sessStdOut, err := sess.StdoutPipe()
	if err != nil {
		return &k3pi.Result{}, err
	}

	output := bytes.Buffer{}

	wg := sync.WaitGroup{}

	var stdOutWriter io.Writer
	if s.writeToStdOut {
		stdOutWriter = io.MultiWriter(os.Stdout, &output)
	} else {
		stdOutWriter = io.MultiWriter(&output)
	}

	wg.Add(1)
	go func() {
		io.Copy(stdOutWriter, sessStdOut)
		wg.Done()
	}()
	sessStderr, err := sess.StderrPipe()
	if err != nil {
		return &k3pi.Result{}, err
	}

	errorOutput := bytes.Buffer{}
	stdErrWriter := io.MultiWriter(os.Stderr, &errorOutput)
	wg.Add(1)
	go func() {
		io.Copy(stdErrWriter, sessStderr)
		wg.Done()
	}()

	err = sess.Run(command)

	wg.Wait()

	if err != nil {
		return &k3pi.Result{}, err
	}

	return &k3pi.Result{
		StdErr: errorOutput.Bytes(),
		StdOut: output.Bytes(),
	}, nil
}

func NewCmdOperator(address string, config *ssh.ClientConfig, writeToStdOut bool) (k3pi.CmdOperator, error) {
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, err
	}

	cmdOperator := cmdRunner{
		writeToStdOut: writeToStdOut,
		client: client,
	}

	return &cmdOperator, nil
}

func sshAgent(publicKeyPath string) (ssh.AuthMethod, func() error) {
	if sshAgentConn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		sshAgent := agent.NewClient(sshAgentConn)

		keys, _ := sshAgent.List()
		if len(keys) == 0 {
			return nil, sshAgentConn.Close
		}

		pubkey, err := ioutil.ReadFile(publicKeyPath)
		if err != nil {
			return nil, sshAgentConn.Close
		}

		authkey, _, _, _, err := ssh.ParseAuthorizedKey(pubkey)
		if err != nil {
			return nil, sshAgentConn.Close
		}
		parsedkey := authkey.Marshal()

		for _, key := range keys {
			if bytes.Equal(key.Blob, parsedkey) {
				return ssh.PublicKeysCallback(sshAgent.Signers), sshAgentConn.Close
			}
		}
	}
	return nil, func() error { return nil }
}

