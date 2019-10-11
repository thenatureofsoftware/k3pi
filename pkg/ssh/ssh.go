/*
Copyright © 2019 The Nature of Software Nordic AB <lars@thenatureofsoftware.se>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package ssh

import (
	"bytes"
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
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

		sshAgent, sshAgentCloseHandler := sshAgent(keyPath + ".pub")
		if sshAgent != nil {
			return sshAgent, sshAgentCloseHandler, nil
		}

		defer sshAgentCloseHandler()

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
		return nil, nil, errors.Wrap(err, fmt.Sprintf("unable to load the ssh key from path %q", settings.GetKeyPath()))
	}

	return &ssh.ClientConfig{
		User: settings.User,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * 3,
	}, closeSSHAgent, nil
}

func NewClientConfigFor(node *model.Node) (*ssh.ClientConfig, func() error, error) {
	auth := node.Auth
	if auth.Type == "ssh-key" {
		config, closeHandler, err := NewClientConfig(&Settings{
			User:    auth.User,
			KeyPath: auth.SSHKey,
			Port:    "22",
		})
		if err != nil {
			return nil, nil, err
		}
		return config, closeHandler, nil
	} else {
		config, closeHandler := PasswordClientConfig(auth.User, auth.Password)
		return config, closeHandler, nil
	}
}

func PasswordClientConfig(username string, password string) (*ssh.ClientConfig, func() error) {
	return &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * 3,
	}, func() error { return nil }
}

type cmdRunner struct {
	writeToStdOut bool
	client        *ssh.Client
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
		_, _ = io.Copy(stdOutWriter, sessStdOut)
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
		_, _ = io.Copy(stdErrWriter, sessStderr)
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

type dryRunCmdRunner struct {
}

func (d dryRunCmdRunner) Close() error {
	return nil
}

func (d dryRunCmdRunner) Execute(command string) (*Result, error) {
	//fmt.Printf("%s\n", command)
	return &Result{
		StdOut: []byte("\n"),
		StdErr: []byte{},
	}, nil
}

func NewCmdOperator(ctx *CmdOperatorCtx) (CmdOperator, error) {
	client, err := ssh.Dial("tcp", ctx.Address.String(), ctx.SSHClientConfig)
	if err != nil {
		return nil, err
	}

	cmdOperator := cmdRunner{
		writeToStdOut: ctx.EnableStdOut,
		client:        client,
	}

	return &cmdOperator, nil
}

func NewDryRunCmdOperator(ctx *CmdOperatorCtx) (CmdOperator, error) {
	client, err := ssh.Dial("tcp", ctx.Address.String(), ctx.SSHClientConfig)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to connect to %s", ctx.Address))
	}
	defer client.Close()

	return &dryRunCmdRunner{}, nil
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
