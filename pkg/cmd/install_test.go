package cmd

import (
	"github.com/TheNatureOfSoftware/k3pi/pkg"
	"github.com/TheNatureOfSoftware/k3pi/pkg/misc"
	"gopkg.in/yaml.v2"
	"os"
	"testing"
)

var nodeYaml = `
hostname: black-pearl
address: 127.0.0.1
arch: armv7l
auth:
  type: basic-auth
  user: pirate
  password: hypriot
`

func TestMakeInstaller(t *testing.T) {
	node := pkg.Node{
		Address: "0.0.0.0",
		Auth:    pkg.Auth{},
		Arch:    "aarch64",
	}

	server := pkg.K3sTarget{
		SSHAuthorizedKeys: []string{},
		Node:              &node,
	}
	server.Node.Address = "192.168.1.10"

	agentAddresses := []string{"192.168.1.11", "192.168.1.12", "192.168.1.13"}
	agents := []pkg.K3sTarget{}
	for _, v := range agentAddresses {
		agent := node
		agent.Address = v
		agents = append(agents, pkg.K3sTarget{
			SSHAuthorizedKeys: []string{},
			Node:              &agent,
		})
	}

	task := &InstallTask{
		DryRun: false,
		Server: &server,
		Agents: &agents,
	}
	installers := MakeInstallers(task)

	want := 4
	if count := len(*installers); count != want {
		t.Errorf("expected %d installers, got %d", want, count)
	}
}

func TestInstaller_Install(t *testing.T) {
	//t.Skip("manual test")
	node := &pkg.Node{}
	err := yaml.Unmarshal([]byte(nodeYaml), node)
	misc.CheckError(err, "failed to load node from yaml string")

	node.Address = "192.168.1.128"
	node.Hostname = "k3pi-1"
	server := pkg.K3sTarget{
		SSHAuthorizedKeys: []string{
			"github:larmog",
		},
		ServerIP: "192.168.1.126",
		Node: node,
	}

	task := &InstallTask{
		DryRun: false,
		Server: &server,
		Agents: &[]pkg.K3sTarget{},
	}

	resourceDir := MakeResourceDir(task)
	defer os.RemoveAll(resourceDir)

	installer := makeInstaller(task, &server, false)

	installer.Install(resourceDir)
}
