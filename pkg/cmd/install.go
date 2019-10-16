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

// Package cmd handles k3pi use cases
package cmd

import (
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg/client"
	"github.com/TheNatureOfSoftware/k3pi/pkg/install"
	"github.com/TheNatureOfSoftware/k3pi/pkg/misc"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
	"net"
	"os"
	"strings"
	"time"
)

const (
	// K3OSDefaultSSHAuthorizedKey is the default authorized key to be included in k3OS config
	K3OSDefaultSSHAuthorizedKey = "~/.ssh/id_rsa.pub"
)

// InstallArgs is a parameter type for calling install function
type InstallArgs struct {
	model.Nodes
	model.SSHKeys
	Token, ServerID string
	*install.HostnameSpec
	DryRun, Confirmed bool
	Templates         *install.ConfigTemplates
	K3OSVersion       string
}

// Install installs k3os on all nodes.
func Install(args *InstallArgs) error {

	generateHostname(args.Nodes, args.HostnameSpec)

	serverNode, agentNodes, err := SelectServerAndAgents(args.Nodes, args.ServerID)
	misc.PanicOnError(err, "failed to resolve server and agents")

	if serverNode != nil {
		misc.Info(fmt.Sprintf("Server:\t%s (%s)", serverNode.Hostname, serverNode.Address))
	} else {
		if len(args.Token) == 0 {
			return fmt.Errorf("no server selected and no join token")
		}
	}

	token := args.Token
	if len(token) == 0 {
		token = misc.GenerateToken()
	}

	misc.Info(fmt.Sprintf("Agents:\t%s", agentNodes.Info(func(n *model.Node) string {
		return fmt.Sprintf("%s (%s)", n.Hostname, n.Address)
	})))

	if !args.Confirmed {
		if misc.DataPipedIn() {
			return fmt.Errorf("install needs to be confirmed (--yes|-y)")
		}
		fmt.Printf("Overwrire all nodes? (y/N): ")
		var reply string
		_, _ = fmt.Scanln(&reply)
		if answer := strings.TrimSpace(strings.ToUpper(string(reply))); answer != "YES" && answer != "Y" {
			return nil
		}
	}

	var serverTarget *model.K3OSNode
	agentTargets := model.NewK3OSNodes(agentNodes, args.SSHKeys, token)

	if serverNode != nil {
		serverTarget = model.NewK3OSNode(serverNode, args.SSHKeys, token)
		agentTargets.SetServerIP(serverNode.Address.IP)
	} else {
		serverIP := net.ParseIP(args.ServerID)
		if serverIP == nil {
			return fmt.Errorf("no server node found and --server '%s' is not a valid IP address", args.ServerID)
		}
		agentTargets.SetServerIP(serverIP.String())
	}

	installTask := &install.OSInstallTask{
		OSImageTask: install.OSImageTask{
			Task: model.Task{
				DryRun: args.DryRun,
			},
			Version:       args.K3OSVersion,
			ClientFactory: client.NewClientFactory(),
		},
		Server:    serverTarget,
		Agents:    agentTargets,
		Templates: args.Templates,
	}

	resourceDir := install.MakeResourceDir(installTask)
	defer os.RemoveAll(resourceDir)

	factory := installerFactories.GetFactory(installTask)
	if factory == nil {
		return fmt.Errorf("installer factory not found for task: %T", installTask)
	}

	installers := factory.MakeInstallers(installTask, resourceDir)

	err = install.Run(installers)
	if err != nil {
		return err
	}

	if serverNode != nil && !args.DryRun {

		serverNode.Auth = model.Auth{
			Type:   model.AuthTypeSSHKey,
			User:   "rancher",
			SSHKey: "~/.ssh/id_rsa",
		}

		if err = install.WaitForNode(client.NewClientFactory(), serverNode, time.Second*60); err == nil {

			var waitForNodeErr error
			fmt.Printf("Waiting for kubeconfig ... ")
			fn := misc.CreateTempFilename(".", "k3s-*.yaml")

			for i := 0; i < 12; i++ {
				waitForNodeErr = misc.CopyKubeconfig(fn, serverNode)
				if waitForNodeErr != nil {
					time.Sleep(time.Second * 15)
				} else {
					fmt.Printf(" OK\n")
					fmt.Printf(" Saved to: %s\n", fn)
					break
				}
			}
			if waitForNodeErr != nil {
				fmt.Printf(" Failed\n")
				return waitForNodeErr
			}
		} else {
			return err
		}
	}

	return nil
}

func generateHostname(nodes model.Nodes, spec *install.HostnameSpec) {
	for i, n := range nodes {
		n.Hostname = spec.GetHostname(i + 1)
	}
}

// SelectServerAndAgents selects the server and returns server and agents separated
func SelectServerAndAgents(nodes model.Nodes, serverID string) (*model.Node, model.Nodes, error) {

	var serverNode *model.Node = nil
	var agentNodes model.Nodes

	for _, node := range nodes {
		if node.Hostname == serverID || node.Address.IP == serverID {
			serverNode = node
		} else {
			agentNodes = append(agentNodes, node)
		}
	}

	return serverNode, agentNodes, nil
}
