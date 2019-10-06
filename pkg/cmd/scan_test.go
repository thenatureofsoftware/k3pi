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
package cmd

import (
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg"
	"github.com/TheNatureOfSoftware/k3pi/pkg/ssh"
	"testing"
)

type mockHostScanner struct {
	returnError bool
}

type MockCmdOperator struct {
	Results map[string]pkg.Result
}

func (op MockCmdOperator) Close() error {
	return nil
}

func (op MockCmdOperator) Execute(command string) (*pkg.Result, error) {
	if result, ok := op.Results[command]; ok {
		return &result, nil
	} else {
		return &result, fmt.Errorf("command not found")
	}
}

func createMockCmdOperator(ctx *pkg.CmdOperatorCtx) (pkg.CmdOperator, error) {
	return MockCmdOperator{Results: make(map[string]pkg.Result)}, nil
}

func (s mockHostScanner) ScanForAliveHosts(cidr string) (*[]string, error) {
	if s.returnError {
		return nil, fmt.Errorf("failed to scan for hosts with CIDR: %s", cidr)
	}
	return &[]string{"127.0.0.1"}, nil
}

func TestScanForRaspberries(t *testing.T) {
	cmdOpFactory := &pkg.CmdOperatorFactory{Create: createMockCmdOperator}
	scanRequest := &ScanRequest{
		Cidr:              "127.0.0.1/32",
		HostnameSubString: "",
		SSHSettings: &ssh.Settings{
			User:    "",
			KeyPath: "~/.ssh/id_rsa",
			Port:    "22",
		},
		UserCredentials: make(map[string]string),
	}
	ScanForRaspberries(scanRequest, &mockHostScanner{}, cmdOpFactory)
}
