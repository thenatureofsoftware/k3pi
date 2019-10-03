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

type installer struct {
	resourceDir     string
	config          *[]byte
	target          *pkg.Target
	operatorFactory *pkg.CmdOperatorFactory
}

func (ins *installer) Install() error {
	sshConfig, sshAgentCloseHandler := ssh.NewClientConfigFor(ins.target.Node)
	defer sshAgentCloseHandler()

	address := ins.target.Node.Address
	sshAddress := fmt.Sprintf("%s:%d", address, 22)

	scpClient := scp.NewClient(sshAddress, sshConfig)
	err := scpClient.Connect()
	misc.CheckError(err, fmt.Sprintf("scp client failed to connect to %s", address))

	imageFile, err := os.Open(ins.target.GetImageFilePath(ins.resourceDir))
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

func MakeInstallers(task *pkg.InstallTask, resourceDir string) pkg.Installers {

	var installers pkg.Installers

	installers = append(installers, makeInstaller(task, task.Server, resourceDir, true))

	for _, agent := range task.Agents {
		installers = append(installers, makeInstaller(task, agent, resourceDir, false))
	}

	return installers
}

func MakeResourceDir(task *pkg.InstallTask) string {
	home, err := homedir.Dir()
	misc.CheckError(err, "failed to resolve home directory")

	resourceDir, err := ioutil.TempDir(home, ".k3pi-")
	misc.CheckError(err, "failed to create resource directory")

	images := make(map[string]string)
	images[task.Server.GetImageFilename()] = fmt.Sprintf(checkSumFileTemplate, task.Server.Node.GetArch())
	for _, agent := range task.Agents {
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

func makeInstaller(task *pkg.InstallTask, target *pkg.Target, resourceDir string, server bool) pkg.Installer {

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
		resourceDir:     resourceDir,
		config:          configYaml,
		target:          target,
		operatorFactory: cmdOperatorFactory,
	}
}

// Installs k3os on all nodes.
func Install(nodes pkg.Nodes, sshKeys []string, serverId string, token string, dryRun bool) error {

	serverNode, agentNodes, err := SelectServerAndAgents(nodes, serverId)
	misc.CheckError(err, "failed to resolve server and agents")

	if serverNode != nil {
		fmt.Printf("server:\t%s\n", serverNode.Address)
	} else {
		if len(token) == 0 {
			return fmt.Errorf("no server selected and no join token")
		}
	}

	fmt.Printf("agents:\t%s\n", agentNodes.IPAddresses())

	var serverTarget *pkg.Target
	agentTargets := agentNodes.Targets(sshKeys)

	if serverNode != nil {
		serverTarget = serverNode.GetTarget(sshKeys)
		agentTargets.SetServerIP(serverNode.Address)
	}

	installTask := &pkg.InstallTask{
		DryRun: dryRun,
		Server: serverTarget,
		Agents: agentTargets,
	}

	resourceDir := MakeResourceDir(installTask)
	defer os.RemoveAll(resourceDir)

	installers := MakeInstallers(installTask, resourceDir)

	err = runInstall(installers)
	if err != nil {
		return err
	}

	return nil
}

type installResult struct {
	installer pkg.Installer
	err       error
}

func runInstall(installers pkg.Installers) error {
	installChan := make(chan pkg.Installer, 5)
	doneChan := make(chan installResult)

	for i := 0; i < 5; i++ {
		go func() {
			installer := <-installChan
			err := installer.Install()
			doneChan <- installResult{
				installer: installer,
				err:       err,
			}
		}()
	}

	for _, installer := range installers {
		installChan <- installer
	}

	var installErrors []error
	for i := 0; i < len(installers); i++ {
		result := <-doneChan
		if result.err != nil {
			installErrors = append(installErrors, errors.Wrap(result.err, fmt.Sprintf("install failed for installer: %v", result.installer)))
		}
	}

	if len(installErrors) > 0 {
		return fmt.Errorf("install errors: %s", installErrors)
	} else {
		return nil
	}
}

func SelectServerAndAgents(nodes pkg.Nodes, serverId string) (*pkg.Node, pkg.Nodes, error) {

	var serverNode *pkg.Node = nil
	var agentNodes pkg.Nodes

	for _, node := range nodes {
		if node.Hostname == serverId || node.Address == serverId {
			serverNode = node
		} else {
			agentNodes = append(agentNodes, node)
		}
	}

	return serverNode, agentNodes, nil
}
