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
	"github.com/TheNatureOfSoftware/k3pi/pkg"
	"github.com/TheNatureOfSoftware/k3pi/pkg/misc"
	"github.com/pkg/errors"
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
	var agents []pkg.K3sTarget
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
	t.Skip("manual test")
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

	_ = installer.Install(resourceDir)
}

func TestSelectServerAndAgents_No_Match(t *testing.T) {
	nodes := []*pkg.Node{{}, {}, {}, {}}
	server, agents, err := SelectServerAndAgents(nodes, "missing")

	if err != nil {
		t.Error(errors.Wrap(err,"unexpected error"))
	}

	if server != nil {
		t.Errorf("no server expected")
	}

	expectedAgentCount := 4
	if actual := len(agents); actual != expectedAgentCount {
		t.Errorf("expected %d agents, actual: %d", expectedAgentCount, actual)
	}
}

func TestSelectServerAndAgents_No_Nodes(t *testing.T) {
	var nodes []*pkg.Node
	server, agents, err := SelectServerAndAgents(nodes, "my-server")

	if err != nil {
		t.Error(errors.Wrap(err,"unexpected error"))
	}

	if server != nil {
		t.Errorf("no server expected")
	}

	expectedAgentCount := 0
	if actual := len(agents); actual != expectedAgentCount {
		t.Errorf("expected %d agents, actual: %d", expectedAgentCount, actual)
	}
}

func TestSelectServerAndAgents_Match_Hostname(t *testing.T) {
	hostname := "my-server"
	nodes := []*pkg.Node{{}, {}, {Hostname: hostname}, {}}
	server, agents, err := SelectServerAndAgents(nodes, hostname)

	if err != nil {
		t.Error(errors.Wrap(err,"unexpected error"))
	}

	if server == nil {
		t.Errorf("expected server is nil")
	}

	expectedAgentCount := 3
	if actual := len(agents); actual != expectedAgentCount {
		t.Errorf("expected %d agents, actual: %d", expectedAgentCount, actual)
	}
}

func TestSelectServerAndAgents_Match_Address(t *testing.T) {
	address := "my-server"
	nodes := []*pkg.Node{{Address: address}}
	server, agents, err := SelectServerAndAgents(nodes, address)

	if err != nil {
		t.Error(errors.Wrap(err,"unexpected error"))
	}

	if server == nil {
		t.Errorf("expected server is nil")
	}

	expectedAgentCount := 0
	if actual := len(agents); actual != expectedAgentCount {
		t.Errorf("expected %d agents, actual: %d", expectedAgentCount, actual)
	}
}
