package install

import (
	"github.com/TheNatureOfSoftware/k3pi/pkg/client"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
	"github.com/TheNatureOfSoftware/k3pi/test"
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

	testNodes := test.CreateNodes()

	server := model.K3OSNode{
		SSHAuthorizedKeys: []string{},
		Node:              *testNodes[0],
	}

	agentNodes := testNodes[1:]
	var agents model.K3OSNodes
	for _, v := range agentNodes {
		agent := *v
		agents = append(agents, &model.K3OSNode{
			SSHAuthorizedKeys: []string{},
			Node:              agent,
		})
	}

	cf, _ := client.NewFakeClientFactory()
	task := &OSInstallTask{
		OSImageTask: OSImageTask{
			Task:          model.Task{},
			Version:       model.DefaultK3OSVersion,
			ClientFactory: cf,
		},
		Server:    &server,
		Agents:    agents,
		Templates: &ConfigTemplates{},
	}

	resourceDir := MakeResourceDir(task)
	defer os.RemoveAll(resourceDir)

	installers := OSInstallerFactory{}.MakeInstallers(task, resourceDir)

	want := len(testNodes)
	if count := len(installers); count != want {
		t.Errorf("expected %d installers, got %d", want, count)
	}
}

func TestOSInstaller_Install(t *testing.T) {
	node := test.CreateNodes()[0]
	node.Hostname = "k3pi-1"

	server := model.K3OSNode{
		SSHAuthorizedKeys: []string{
			"github:larmog",
		},
		ServerIP: "192.168.1.126",
		Node:     *node,
	}

	cf, _ := client.NewFakeClientFactory()
	task := &OSInstallTask{
		OSImageTask: OSImageTask{
			Task:          model.Task{},
			Version:       model.DefaultK3OSVersion,
			ClientFactory: cf,
		},
		Server:    &server,
		Agents:    model.K3OSNodes{},
		Templates: &ConfigTemplates{},
	}

	resourceDir := MakeResourceDir(task)
	defer os.RemoveAll(resourceDir)

	installer := makeInstaller(task, &server, resourceDir, false)

	_ = installer.Install()
}
