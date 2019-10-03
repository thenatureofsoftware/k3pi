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

var imageFilenameTmpl = "k3os-rootfs-%s.tar.gz"

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

func (n *Node) GetK3sTarget(sshAuthorizedKeys []string) *K3sTarget {
	return &K3sTarget{
		SSHAuthorizedKeys: sshAuthorizedKeys,
		Node:              n,
	}
}

type Auth struct {
	Type     string `json:"type"`
	User     string `json:"user"`
	Password string `json:"password,omitempty"`
	SSHKey   string `json:"ssh_key,omitempty"`
}

type K3sTarget struct {
	SSHAuthorizedKeys []string
	ServerIP string
	Node              *Node
}

func (target *K3sTarget) GetImageFilename() string {
	return fmt.Sprintf(imageFilenameTmpl, target.Node.GetArch())
}

func (target *K3sTarget) GetImageFilePath(resourceDir string) string {
	return fmt.Sprintf("%s%s%s", resourceDir, string(os.PathSeparator), target.GetImageFilename())
}
