package pkg

import "golang.org/x/crypto/ssh"

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
	SSHKey   string `json:"ssh-key,omitempty"`
}

type K3sTarget struct {
	SSHAuthorizedKeys []string
	Node              *Node
}
