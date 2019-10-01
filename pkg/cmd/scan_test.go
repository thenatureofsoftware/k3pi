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
    return MockCmdOperator{Results:make(map[string]pkg.Result)}, nil
}

func (s mockHostScanner) ScanForAliveHosts(cidr string) (*[]string, error) {
    if s.returnError {
        return nil, fmt.Errorf("failed to scan for hosts with CIDR: %s", cidr)
    }
    return &[]string {"127.0.0.1"}, nil
}

func TestScanForRaspberries(t *testing.T) {
    cmdOpFactory := &pkg.CmdOperatorFactory{Create: createMockCmdOperator}
    scanRequest := &ScanRequest{
        Cidr:              "127.0.0.1/32",
        HostnameSubString: "",
        SSHSettings:       &ssh.Settings{
            User:    "",
            KeyPath: "~/.ssh/id_rsa",
            Port:    "22",
        },
        UserCredentials:   make(map[string]string),
    }
    ScanForRaspberries(scanRequest, &mockHostScanner{}, cmdOpFactory)
}
