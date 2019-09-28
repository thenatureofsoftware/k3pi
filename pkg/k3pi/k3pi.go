package k3pi

// The stdin and stdout from executing a command.
type Result struct {
    StdOut []byte
    StdErr []byte
}

type CmdOperator interface {
    Close() error
    Execute(command string) (*Result, error)
}

type Member struct {
    IP string `yaml:"ip_address"`
    Type string `yaml:"type"`
}
