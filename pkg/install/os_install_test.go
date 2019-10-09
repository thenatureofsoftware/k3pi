package install

import (
	"github.com/TheNatureOfSoftware/k3pi/pkg"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
	"github.com/TheNatureOfSoftware/k3pi/pkg/misc"
	"github.com/kubernetes-sigs/yaml"
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

func TestOSInstallerFactory_MakeInstallers(t *testing.T) {

	node := model.Node{
		Address: "0.0.0.0",
		Auth:    model.Auth{},
		Arch:    "aarch64",
	}

	server := model.K3OSNode{
		SSHAuthorizedKeys: []string{},
		Node:              node,
	}

	server.Node.Address = "192.168.1.10"

	agentAddresses := []string{"192.168.1.11", "192.168.1.12", "192.168.1.13"}
	var agents model.K3OSNodes
	for _, v := range agentAddresses {
		agent := node
		agent.Address = v
		agents = append(agents, &model.K3OSNode{
			SSHAuthorizedKeys: []string{},
			Node:              agent,
		})
	}

	task := NewOSInstallTask(&server, agents, &pkg.ConfigTemplates{}, false)

	resourceDir := MakeResourceDir(task)
	defer os.RemoveAll(resourceDir)

	installers := OSInstallerFactory{}.MakeInstallers(task, resourceDir)

	want := 4
	if count := len(installers); count != want {
		t.Errorf("expected %d installers, got %d", want, count)
	}
}

func TestOSInstaller_Install(t *testing.T) {
	t.Skip("manual test")
	node := &model.Node{}
	err := yaml.Unmarshal([]byte(nodeYaml), node)
	misc.PanicOnError(err, "failed to load node from yaml string")

	node.Address = "192.168.1.128"
	node.Hostname = "k3pi-1"
	server := model.K3OSNode{
		SSHAuthorizedKeys: []string{
			"github:larmog",
		},
		ServerIP: "192.168.1.126",
		Node:     *node,
	}

	task := &OSInstallTask{
		Server: &server,
		Agents: model.K3OSNodes{},
		Task: model.Task{
			DryRun: false,
		},
	}

	resourceDir := MakeResourceDir(task)
	defer os.RemoveAll(resourceDir)

	installer := makeInstaller(task, &server, resourceDir, false)

	_ = installer.Install()
}
