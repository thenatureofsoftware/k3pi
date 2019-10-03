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
	"bufio"
	"bytes"
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg"
	"github.com/TheNatureOfSoftware/k3pi/pkg/config"
	"github.com/TheNatureOfSoftware/k3pi/pkg/misc"
	"github.com/TheNatureOfSoftware/k3pi/pkg/ssh"
	"github.com/bramvdbogaerde/go-scp"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
)

var checkSumFileTemplate = "sha256sum-%s.txt"

type Installer interface {
	Install(resourceDir string) error
}

type InstallTask struct {
	DryRun            bool
	Server            *pkg.K3sTarget
	Agents            *[]pkg.K3sTarget
	SSHAuthorizedKeys []string
}

type installer struct {
	resourceDir     string
	config          *[]byte
	target          *pkg.K3sTarget
	operatorFactory *pkg.CmdOperatorFactory
}

func (ins *installer) Install(resourceDir string) error {
	sshConfig, sshAgentCloseHandler := ssh.NewClientConfigFor(ins.target.Node)
	defer sshAgentCloseHandler()

	address := ins.target.Node.Address
	sshAddress := fmt.Sprintf("%s:%d", address, 22)

	scpClient := scp.NewClient(sshAddress, sshConfig)
	err := scpClient.Connect()
	misc.CheckError(err, fmt.Sprintf("scp client failed to connect to %s", address))

	imageFile, err := os.Open(ins.target.GetImageFilePath(resourceDir))
	misc.CheckError(err, "failed to open image file")
	defer imageFile.Close()
	stat, err := imageFile.Stat()
	misc.CheckError(err, "failed to get file info")

	err = scpClient.Copy(bufio.NewReader(imageFile), fmt.Sprintf("~/%s", ins.target.GetImageFilename()), "0655", stat.Size())
	misc.CheckError(err, "failed to copy image file")

	// It's strange but we need to close and open for each file
	_ = scpClient.Session.Close()
	err = scpClient.Connect()
	misc.CheckError(err, fmt.Sprintf("scp client failed to connect to %s", address))
	defer scpClient.Session.Close()

	err = scpClient.Copy(bytes.NewReader(*ins.config), fmt.Sprintf("~/%s", "config.yaml"), "0655", int64(len(*ins.config)))
	misc.CheckError(err, "failed to copy config file")

	ctx := &pkg.CmdOperatorCtx{
		Address:         sshAddress,
		SSHClientConfig: sshConfig,
		EnableStdOut:    false,
	}

	operator, err := ins.operatorFactory.Create(ctx)
    misc.CheckError(err, fmt.Sprintf("failed to connect to %s", ctx.Address))

	result, err := operator.Execute(fmt.Sprintf("sudo tar zxvf %s --strip-components=1 -C /", ins.target.GetImageFilename()))
	if err2 := errors.Wrap(err, fmt.Sprintf("failed to extract %s, result:\n %v", ins.target.GetImageFilename(), result)); err2 != nil {
		return err2
	}

	result, err = operator.Execute("sudo cp config.yaml /k3os/system/config.yaml && sudo sync && sudo reboot -f")
	if err2 := errors.Wrap(err, fmt.Sprintf("failed to install config and reboot:\n %v", result)); err2 != nil {
		return err2
	}

	return nil
}

func MakeInstallers(task *InstallTask) *[]Installer {
	fmt.Printf("Installing %s as server and %d agents\n", task.Server.Node.Hostname, len(*task.Agents))

	resourceDir := MakeResourceDir(task)
	defer os.RemoveAll(resourceDir)

	installers := []Installer{}

	installers = append(installers, makeInstaller(task, task.Server, true))

	for _, agent := range *task.Agents {
		installers = append(installers, makeInstaller(task, &agent, false))
	}

	return &installers
}

func MakeResourceDir(task *InstallTask) string {
	homedir, err := homedir.Dir()
	misc.CheckError(err, "failed to resolve home directory")

	resourceDir, err := ioutil.TempDir(homedir, ".k3pi-")
    misc.CheckError(err, "failed to create resource directory")

	images := make(map[string]string)
	images[task.Server.GetImageFilename()] = fmt.Sprintf(checkSumFileTemplate, task.Server.Node.GetArch())
	for _, agent := range *task.Agents {
		images[agent.GetImageFilename()] = fmt.Sprintf(checkSumFileTemplate, agent.Node.GetArch())
	}

	url := "https://github.com/rancher/k3os/releases/download/v0.3.0/%s"
	pathSeparator := string(os.PathSeparator)
	for imageFile, checkSumFile := range images {
		download := misc.FileDownload{
			Filename:         fmt.Sprintf("%s%s%s", resourceDir, pathSeparator, imageFile),
			CheckSumFilename: fmt.Sprintf("%s%s%s", resourceDir, pathSeparator, checkSumFile),
			Url:              fmt.Sprintf(url, imageFile),
			CheckSumUrl:      fmt.Sprintf(url, checkSumFile),
		}
		err := misc.DownloadAndVerify(download)
        misc.CheckError(err, "failed to create resource directory")
	}

	return resourceDir
}

func makeInstaller(task *InstallTask, target *pkg.K3sTarget, server bool) Installer {

	var configYaml *[]byte
	var err error

	if server {
		configYaml, err = config.NewServerConfig("", target)
	} else {
		configYaml, err = config.NewAgentConfig("", target)
	}

	misc.CheckError(err, "failed to create server installer")

	cmdOperatorFactory := &pkg.CmdOperatorFactory{}
	if task.DryRun {
		cmdOperatorFactory.Create = ssh.NewDryRunCmdOperator
	} else {
		cmdOperatorFactory.Create = ssh.NewCmdOperator
	}

	return &installer{
		config:          configYaml,
		target:          target,
		operatorFactory: cmdOperatorFactory,
	}
}

// Installs k3os on all nodes.
func Install(nodes []*pkg.Node, sshKeys []string, serverId string, token string, dryRun bool) error {

    serverNode, agentNodes, err := SelectServerAndAgents(nodes, serverId)
    misc.CheckError(err, "failed to resolve server and agents")

    if serverNode != nil {
        fmt.Printf("server:\t%s\n", serverNode.Address)
    } else {
        if len(token) == 0 {
            return fmt.Errorf("no server selected and no join token")
        }
    }

    var agentIPs []string
    for _, v := range agentNodes { agentIPs = append(agentIPs, v.Address) }
    fmt.Printf("agents:\t%s\n", agentIPs)

    return nil
}

func SelectServerAndAgents(nodes []*pkg.Node, serverId string) (*pkg.Node, []*pkg.Node, error) {
    var serverNode *pkg.Node = nil
    var agentNodes []*pkg.Node

    for _, node := range nodes {
        if node.Hostname == serverId || node.Address == serverId {
            serverNode = node
        } else {
            agentNodes = append(agentNodes, node)
        }
    }
    return serverNode, agentNodes, nil
}


