package client

import (
	"bytes"
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg/misc"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
	"io"
	"io/ioutil"
)

func NewFakeClient(auth *model.Auth, address *model.Address) (Client, error) {
	return &fakeClient{
		auth:    auth,
		address: address,
	}, nil
}

type fakeClient struct {
	auth    *model.Auth
	address *model.Address
}

func (f *fakeClient) CopyReader(reader io.Reader, remotePath string) error {
	b, err := ioutil.ReadAll(reader)
	misc.PanicOnError(err, "failed to read content")

	fmt.Printf("scp %s:\n%s\n", remotePath, b)

	return nil
}

func (f *fakeClient) Copy(filename, remotePath string) error {
	fmt.Printf("scp -i %s -P %d -o StrictHostKeyChecking=no %s %s\n",
		f.auth.SSHKey,
		f.address.Port,
		filename,
		fmt.Sprintf("%s@%s:%s", f.auth.User, f.address.IP, remotePath))
	return nil
}

func (f *fakeClient) Cmd(cmd string) Script {
	fs := &fakeScript{}
	return fs.Cmd(cmd)
}

func (f *fakeClient) Cmdf(cmd string, a ...interface{}) Script {
	return f.Cmd(fmt.Sprintf(cmd, a...))
}

type fakeScript struct {
	runError    error
	commands    []string
	output      string
	outputError error
}

func (s *fakeScript) Cmd(cmd string) Script {
	s.commands = append(s.commands, cmd)
	return s
}

func (s *fakeScript) Cmdf(cmd string, a ...interface{}) Script {
	return s.Cmd(fmt.Sprintf(cmd, a...))
}

func (s *fakeScript) Run() error {
	return s.runError
}

func (s *fakeScript) Output() ([]byte, error) {
	var out []byte
	buffer := bytes.NewBuffer(out)
	buffer.WriteString(fmt.Sprintf("Commands:\n-----------------------------------\n"))
	for _, cmd := range s.commands {
		buffer.WriteString(fmt.Sprintf("\t%s\n", cmd))
	}
	buffer.WriteString(fmt.Sprintln("-----------------------------------"))
	fmt.Println(buffer.String())
	return []byte(s.output), s.outputError
}
