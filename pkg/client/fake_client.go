package client

import (
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
	"strings"
	"sync"
)

// NewFakeClientFactory creates a fake client factory
func NewFakeClientFactory(configurator ...func(script *FakeScript)) (*Factory, *FakeScript) {
	fs := &FakeScript{m: sync.Mutex{}, Interactions:make(map[string][]string)}
	return &Factory{Create: func(auth *model.Auth, address *model.Address) (i Client, e error) {
		fc := &FakeClient{FakeScript: fs}
		fc.Auth = auth
		fc.Address = address
		for _, it := range configurator { it(fc.FakeScript) }
		return fc, nil
	}}, fs
}

// NewFakeClient factory method for creating a fake node client
func NewFakeClient(auth *model.Auth, address *model.Address) (Client, error) {
	return &FakeClient{
		FakeScript: &FakeScript{m: sync.Mutex{}, Interactions:make(map[string][]string)},
		Auth:    auth,
		Address: address,
	}, nil
}

// FakeClient a fake client for testing
type FakeClient struct {
	Auth       *model.Auth
	Address    *model.Address
	FakeScript *FakeScript
}

// CopyBytes fakes copy of []byte to remote path
func (f *FakeClient) CopyBytes(b *[]byte, remotePath string) error {
	fmt.Printf("scp %s:\n%v\n", remotePath, b)
	return nil
}

// Copy fakes copy of file to remote path
func (f *FakeClient) Copy(filename, remotePath string) error {
	fmt.Printf("scp -i %s -P %d -o StrictHostKeyChecking=no %s %s\n",
		f.Auth.SSHKey,
		f.Address.Port,
		filename,
		fmt.Sprintf("%s@%s:%s", f.Auth.User, f.Address.IP, remotePath))
	return nil
}

// Cmd adds command for fake execution
func (f *FakeClient) Cmd(cmd string) Script {
	return f.FakeScript.Cmd(cmd)
}

// Cmdf adds command for fake execution
func (f *FakeClient) Cmdf(cmd string, a ...interface{}) Script {
	return f.Cmd(fmt.Sprintf(cmd, a...))
}

// FakeScript for fake script execution
type FakeScript struct {
	m            sync.Mutex
	Error        error
	InvokedCmds  []string
	Interactions map[string][]string
}

// Expect what command to expect: stdin and stdout
func (s *FakeScript) Expect(stdin, stdout string) {
	if v, ok := s.Interactions[stdin]; ok {
		out := append(v, stdout)
		s.Interactions[stdin]=out
	} else {
		s.Interactions[stdin] = []string{stdout}
	}
}

// Cmd add command for fake execution
func (s *FakeScript) Cmd(cmd string) Script {
	s.m.Lock()
	defer s.m.Unlock()
	s.InvokedCmds = append(s.InvokedCmds, cmd)
	return s
}

// Cmdf add command for fake execution
func (s *FakeScript) Cmdf(cmd string, a ...interface{}) Script {
	return s.Cmd(fmt.Sprintf(cmd, a...))
}

// Run fakes running command on remote host
func (s *FakeScript) Run() error {
	_, err := s.Output()
	return err
}

// HasOutstandingCmds returns true if there are exepcted commands not invoked
func (s *FakeScript) HasOutstandingCmds() bool {
	for _, v := range s.Interactions {
		if len(v) > 0 {
			return true
		}
	}
	return false
}

// Output fakes running command on remote host and returns configured output
func (s *FakeScript) Output() ([]byte, error) {
	s.m.Lock()
	defer s.m.Unlock()

	var output []string
	for _, v := range s.InvokedCmds {
		var cmdOut string
		if out, ok := s.Interactions[v]; ok {
			size := len(out)
			if size > 0 {
				cmdOut = out[0]
				if size > 1 {
					s.Interactions[v] = out[1:]
				} else {
					delete(s.Interactions, v)
				}
			}
		} else {
			cmdOut = ""
		}
		output = append(output, cmdOut)
		fmt.Printf("$ %s\n%s\n", v, cmdOut)
	}

	b := strings.Builder{}
	for _, s := range output {
		b.WriteString(fmt.Sprintf("%s\n", s))
	}

	s.InvokedCmds = []string{}
	return []byte(b.String()), s.Error
}
