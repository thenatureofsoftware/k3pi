package ssh

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
