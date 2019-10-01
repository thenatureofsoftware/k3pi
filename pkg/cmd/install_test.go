package cmd

import (
	"github.com/TheNatureOfSoftware/k3pi/pkg"
	"testing"
)

func TestMakeInstaller(t *testing.T) {
	node := pkg.Node{
		Address: "0.0.0.0",
		Auth:    pkg.Auth{},
		Arch:    "aarch64",
	}

	server := node
	server.Address = "192.168.1.10"

	agentAddresses := []string{"192.168.1.11", "192.168.1.12", "192.168.1.13"}
	agents := []pkg.Node{}
	for _, v := range agentAddresses {
		agent := node
		agent.Address = v
		agents = append(agents, agent)
	}

	task := &InstallTask{
		DryRun: false,
		Server: server,
		Agents: agents,
	}
	installers := MakeInstaller(task)

	want := 4
	if count := len(*installers); count != want {
		t.Errorf("expected %d installers, got %d", want, count)
	}
}
