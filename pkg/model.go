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
package pkg

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"os"
)

const ( ImageFilenameTmpl = "k3os-rootfs-%s.tar.gz" )

type SSHKeys []string

// The stdin and stdout from executing a command.
type Result struct {
	StdOut []byte
	StdErr []byte
}

type CmdOperatorCtx struct {
	Address         string
	SSHClientConfig *ssh.ClientConfig
	EnableStdOut    bool
}

type CmdOperator interface {
	Close() error
	Execute(command string) (*Result, error)
}

type CmdOperatorFactory struct {
	Create func(ctx *CmdOperatorCtx) (CmdOperator, error)
}

type Node struct {
	Hostname string `json:"hostname"`
	Address  string `json:"address"`
	Auth     Auth   `json:"auth"`
	Arch     string `json:"arch"`
}

type Nodes []*Node

func (n *Node) GetArch() string {
	switch n.Arch {
	case "x86_64":
		return "amd64"
	case "armv6l", "armv7l":
		return "arm"
	case "aarch64":
		return "arm64"
	default:
		return "unknown"
	}
}

func (n *Node) GetTarget(sshAuthorizedKeys []string) *Target {
	return &Target{
		SSHAuthorizedKeys: sshAuthorizedKeys,
		Node:              n,
	}
}

func (nodes *Nodes) IPAddresses() []string {
	var ipAddresses []string
	for _, v := range *nodes {
		ipAddresses = append(ipAddresses, v.Address)
	}
	return ipAddresses
}

func (nodes *Nodes) Info(collect func(*Node) string) []string {
	var info []string
	for _, v := range *nodes {
		info = append(info, collect(v))
	}
	return info
}

type Auth struct {
	Type     string `json:"type"`
	User     string `json:"user"`
	Password string `json:"password,omitempty"`
	SSHKey   string `json:"ssh_key,omitempty"`
}

type Target struct {
	SSHAuthorizedKeys []string
	ServerIP          string
	Node              *Node
}

func (target *Target) GetImageFilename() string {
	return fmt.Sprintf(ImageFilenameTmpl, target.Node.GetArch())
}

func (target *Target) GetImageFilePath(resourceDir string) string {
	return fmt.Sprintf("%s%s%s", resourceDir, string(os.PathSeparator), target.GetImageFilename())
}

type Targets []*Target

func (targets *Targets) SetServerIP(serverIP string) {
	for _, target := range *targets {
		target.ServerIP = serverIP
	}
}

func (nodes *Nodes) Targets(sshAuthorizedKeys []string) Targets {
	var targets Targets
	for _, node := range *nodes {
		targets = append(targets, node.GetTarget(sshAuthorizedKeys))
	}
	return targets
}

type Installer interface {
	Install() error
}

type Installers []Installer

type InstallTask struct {
	DryRun bool
	Server *Target
	Agents Targets
}

type HostnameSpec struct {
	Pattern, Prefix string
}

func (h *HostnameSpec) GetHostname(index int) string {
	return fmt.Sprintf(h.Pattern, h.Prefix, index)
}
