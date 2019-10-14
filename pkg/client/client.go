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

// Package client client supporting installing and upgrading remote k3OS nodes
package client

import (
	"fmt"
	"github.com/TheNatureOfSoftware/go-sshclient"
	"github.com/TheNatureOfSoftware/k3pi/pkg/misc"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
	"github.com/bramvdbogaerde/go-scp"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Factory factory for creating new clients
type Factory struct {
	Create func(auth *model.Auth, address *model.Address) (Client, error)
}

// Client runs commands an copies files
type Client interface {
	Cmd(cmd string) Script
	Cmdf(cmd string, a ...interface{}) Script
	Copy(filename, remotePath string) error
	CopyBytes(b *[]byte, remotePath string) error
}

// Script script for running remote commands
type Script interface {
	Cmd(cmd string) Script
	Cmdf(cmd string, a ...interface{}) Script
	Run() error
	Output() ([]byte, error)
}

// NewClient factory method for creating a new node client
func NewClient(auth *model.Auth, address *model.Address) (Client, error) {

	var sshClient *sshclient.Client
	var err error
	if auth.Type == model.AuthTypeSSHKey {
		privateKeyFile, err := homedir.Expand(auth.SSHKey)
		misc.PanicOnError(err, "failed to expand ssh key file path")
		sshClient, err = sshclient.DialWithKey(address.String(), auth.User, privateKeyFile)
	} else {
		sshClient, err = sshclient.DialWithPasswd(address.String(), auth.User, auth.Password)
	}

	if err != nil {
		return nil, err
	}

	if sshClient == nil {
		return nil, fmt.Errorf("ssh client is nil, maybe because of wrong user")
	}

	c := &client{
		auth:      auth,
		address:   address,
		sshClient: sshClient,
	}

	return c, nil
}

type client struct {
	sshClient *sshclient.Client
	auth      *model.Auth
	address   *model.Address
}

func (c *client) Cmd(cmd string) Script {
	rs := c.sshClient.Cmd(cmd)
	return &script{remoteScript: rs}
}

func (c *client) Cmdf(cmd string, a ...interface{}) Script {
	return c.Cmd(fmt.Sprintf(cmd, a...))
}

func (c *client) Copy(filename, remotePath string) error {

	if c.auth.Type != model.AuthTypeSSHKey {
		_, err := c.copyWithPassword(filename, remotePath)
		return err
	}
	out, err := c.copyWithKey(filename, remotePath)
	if err != nil {
		return fmt.Errorf(strings.TrimSpace(string(out)))
	}

	return err
}

func (c *client) copyWithPassword(filename string, remotePath string) ([]byte, error) {
	config, closeHandler := passwordClientConfig(c.auth.User, c.auth.Password)
	defer closeHandler()

	scpClient := scp.NewClient(c.address.String(), config)
	err := scpClient.Connect()
	misc.PanicOnError(err, fmt.Sprintf("failed to connect to %s", c.address))

	// Open a file
	f, _ := os.Open(filename)

	defer scpClient.Close()
	defer f.Close()

	err = scpClient.CopyFromFile(*f, remotePath, "0655")
	var out []byte

	return out, err
}

func passwordClientConfig(username string, password string) (*ssh.ClientConfig, func() error) {
	return &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * 3,
	}, func() error { return nil }
}

func (c *client) copyWithKey(filename string, remotePath string) ([]byte, error) {
	out, err := exec.Command(
		"scp",
		"-i",
		c.auth.SSHKey,
		"-P",
		fmt.Sprintf("%d", c.address.Port),
		"-o",
		"StrictHostKeyChecking=no",
		filename,
		fmt.Sprintf("%s@%s:%s", c.auth.User, c.address.IP, remotePath)).CombinedOutput()
	return out, err
}

func (c *client) CopyBytes(b *[]byte, remotePath string) error {
	f, err := ioutil.TempFile(os.TempDir(), "k3pi-*")
	misc.PanicOnError(err, "failed to create temp file")

	f.Write(*b)
	f.Close()

	err = c.Copy(f.Name(), remotePath)
	defer os.Remove(f.Name())

	return err
}

type script struct {
	remoteScript *sshclient.RemoteScript
}

func (s *script) Cmd(cmd string) Script {
	s.remoteScript = s.remoteScript.Cmd(cmd)
	return s
}

func (s *script) Cmdf(cmd string, a ...interface{}) Script {
	s.remoteScript = s.remoteScript.Cmd(fmt.Sprintf(cmd, a...))
	return s
}

func (s *script) Run() error {
	return s.remoteScript.Run()
}

func (s *script) Output() ([]byte, error) {
	return s.remoteScript.SmartOutput()
}
