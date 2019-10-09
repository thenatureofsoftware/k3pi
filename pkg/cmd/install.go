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
	"github.com/TheNatureOfSoftware/k3pi/pkg/install"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
	"github.com/TheNatureOfSoftware/k3pi/pkg/misc"
	"net"
	"os"
	"strings"
	"time"
)

const DefaultSSHAuthorizedKey = "~/.ssh/id_rsa.pub"

type InstallArgs struct {
	model.Nodes
	model.SSHKeys
	Token, ServerID string
	*pkg.HostnameSpec
	DryRun, Confirmed bool
	Templates         *pkg.ConfigTemplates
}

// Installs k3os on all nodes.
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
		agentTargets.SetServerIP(serverNode.Address)
	} else {
		serverIP := net.ParseIP(args.ServerID)
		if serverIP == nil {
			return fmt.Errorf("no server node found and --server '%s' is not a valid IP address", args.ServerID)
		}
		agentTargets.SetServerIP(serverIP.String())
	}

	installTask := install.NewOSInstallTask(serverTarget, agentTargets, args.Templates, args.DryRun)

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
		if err = misc.WaitForNode(serverNode, nil, time.Second*60); err == nil {

			var waitForNodeErr error
			fmt.Printf("Waiting for kubeconfig ... ")
			fn := misc.CreateTempFileName(".", "k3s-*.yaml")

			for i := 0; i < 6; i++ {
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

func generateHostname(nodes model.Nodes, spec *pkg.HostnameSpec) {
	for i, n := range nodes {
		n.Hostname = spec.GetHostname(i + 1)
	}
}

func SelectServerAndAgents(nodes model.Nodes, serverId string) (*model.Node, model.Nodes, error) {

	var serverNode *model.Node = nil
	var agentNodes model.Nodes

	for _, node := range nodes {
		if node.Hostname == serverId || node.Address == serverId {
			serverNode = node
		} else {
			agentNodes = append(agentNodes, node)
		}
	}

	return serverNode, agentNodes, nil
}
