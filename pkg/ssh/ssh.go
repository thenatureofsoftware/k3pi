package ssh

import (
	"bytes"
	"fmt"
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

// The stdin and stdout from executing a command.
type Result struct {
	StdOut []byte
	StdErr []byte
}

type CmdOperator interface {
	Close() error
	Execute(command string) (*Result, error)
}

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

type cmdRunner struct {
	writeToStdOut bool
	client *ssh.Client
}

func (s *cmdRunner) Close() error {
	return s.client.Close()
}

func (s *cmdRunner) Execute(command string) (*Result, error) {
	sess, err := s.client.NewSession()
	if err != nil {
		return &Result{}, err
	}

	defer sess.Close()

	sessStdOut, err := sess.StdoutPipe()
	if err != nil {
		return &Result{}, err
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
		return &Result{}, err
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
		return &Result{}, err
	}

	return &Result{
		StdErr: errorOutput.Bytes(),
		StdOut: output.Bytes(),
	}, nil
}

func NewCmdOperator(address string, config *ssh.ClientConfig, writeToStdOut bool) (CmdOperator, error) {
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
