package client

import (
	"bytes"
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
)

// NewFakeClient factory method for creating a fake node client
func NewFakeClient(auth *model.Auth, address *model.Address) (Client, error) {
	return &FakeClient{
		Auth:    auth,
		Address: address,
	}, nil
}

type FakeClient struct {
	Auth    *model.Auth
	Address *model.Address
	FakeScript FakeScript
}

func (f *FakeClient) CopyBytes(b *[]byte, remotePath string) error {
	fmt.Printf("scp %s:\n%s\n", remotePath, b)
	return nil
}

func (f *FakeClient) Copy(filename, remotePath string) error {
	fmt.Printf("scp -i %s -P %d -o StrictHostKeyChecking=no %s %s\n",
		f.Auth.SSHKey,
		f.Address.Port,
		filename,
		fmt.Sprintf("%s@%s:%s", f.Auth.User, f.Address.IP, remotePath))
	return nil
}

func (f *FakeClient) Cmd(cmd string) Script {
	f.FakeScript.Cmds = append(f.FakeScript.Cmds, cmd)
	return f.FakeScript.Cmd(cmd)
}

func (f *FakeClient) Cmdf(cmd string, a ...interface{}) Script {
	return f.Cmd(fmt.Sprintf(cmd, a...))
}

type FakeScript struct {
	Error      error
	Cmds       []string
	CmdResults []string
}

func (s *FakeScript) Cmd(cmd string) Script {
	s.Cmds = append(s.Cmds, cmd)
	return s
}

func (s *FakeScript) Cmdf(cmd string, a ...interface{}) Script {
	return s.Cmd(fmt.Sprintf(cmd, a...))
}

func (s *FakeScript) Run() error {
	return s.Error
}

func (s *FakeScript) Output() ([]byte, error) {
	var out []byte
	buffer := bytes.NewBuffer(out)
	buffer.WriteString(fmt.Sprintf("Cmds:\n-----------------------------------\n"))
	for _, cmd := range s.Cmds {
		buffer.WriteString(fmt.Sprintf("\t%s\n", cmd))
	}
	buffer.WriteString(fmt.Sprintln("-----------------------------------"))
	fmt.Println(buffer.String())

	size := len(s.CmdResults)
	if size > 0 {
		result := s.CmdResults[0]
		if size > 1 {
			s.CmdResults = s.CmdResults[1:]
		}
		return []byte(result), s.Error
	}

	return []byte{}, s.Error
}
