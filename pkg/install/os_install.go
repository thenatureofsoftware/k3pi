package install

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg"
	"github.com/TheNatureOfSoftware/k3pi/pkg/config"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
	"github.com/TheNatureOfSoftware/k3pi/pkg/misc"
	"github.com/TheNatureOfSoftware/k3pi/pkg/ssh"
	"github.com/bramvdbogaerde/go-scp"
	"github.com/pkg/errors"
	"os"
)

type OSInstallTask struct {
	model.Task
	Server    *model.K3OSNode
	Agents    model.K3OSNodes
	Templates *pkg.ConfigTemplates
}

func (task *OSInstallTask) GetRemoteAssets() model.RemoteAssets {
	var allNodes = model.K3OSNodes{}
	var resources model.RemoteAssets

	if task.Server != nil {
		allNodes = append(allNodes, task.Server)
	}

	for _, agent := range task.Agents {
		allNodes = append(allNodes, agent)
	}

	for _, node := range allNodes {
		arch := node.GetArch()
		resources = append(resources, &model.RemoteAsset{
			Filename:         task.GetImageFilename(arch),
			FileUrl:          task.GetImageFileUrl(arch),
			CheckSumFilename: task.GetImageCheckSumFilename(arch),
			CheckSumUrl:      task.GetImageCheckSumUrl(arch),
		})
	}

	return resources
}

func (task *OSInstallTask) GetImageFilePath(resourceDir string, arch string) string {
	return fmt.Sprintf("%s%s%s", resourceDir, pkg.PathSeparator, task.GetImageFilename(arch))
}

func (task *OSInstallTask) GetImageCheckSumFilePath(resourceDir string, arch string) string {
	return fmt.Sprintf("%s%s%s", resourceDir, pkg.PathSeparator, task.GetImageCheckSumFilename(arch))
}

func (task *OSInstallTask) GetImageFilename(arch string) string {
	return fmt.Sprintf(pkg.ImageFilenameTmpl, arch)
}

func (task *OSInstallTask) GetImageCheckSumFilename(arch string) string {
	return fmt.Sprintf(pkg.CheckSumFileTemplate, arch)
}

func (task *OSInstallTask) GetImageFileUrl(arch string) string {
	return fmt.Sprintf(pkg.K3OSReleaseUrlTmpl, "v0.3.0", task.GetImageFilename(arch))
}

func (task *OSInstallTask) GetImageCheckSumUrl(arch string) string {
	return fmt.Sprintf(pkg.K3OSReleaseUrlTmpl, "v0.3.0", task.GetImageCheckSumFilename(arch))
}

func NewOSInstallTask(server *model.K3OSNode, agents model.K3OSNodes, templates *pkg.ConfigTemplates, dryRun bool) *OSInstallTask {
	return &OSInstallTask{
		Task:      model.Task{DryRun: dryRun},
		Server:    server,
		Agents:    agents,
		Templates: templates,
	}
}

type OSInstallerFactory struct{}

func (o OSInstallerFactory) Supports(task interface{}) bool {
	return fmt.Sprintf("%T", task) == fmt.Sprintf("%T", &OSInstallTask{})
}

func (o OSInstallerFactory) MakeInstallers(task interface{}, resourceDir string) model.Installers {
	installTask := task.(*OSInstallTask)
	var installers model.Installers

	if installTask.Server != nil {
		installers = append(installers, makeInstaller(installTask, installTask.Server, resourceDir, true))
	}

	for _, agent := range installTask.Agents {
		installers = append(installers, makeInstaller(installTask, agent, resourceDir, false))
	}

	return installers
}

func makeInstaller(task *OSInstallTask, k3OSNode *model.K3OSNode, resourceDir string, server bool) model.Installer {

	var configYaml *[]byte
	var err error

	if server {
		configYaml, err = config.NewServerConfig(task.Templates.ServerTmpl, k3OSNode)
	} else {
		configYaml, err = config.NewAgentConfig(task.Templates.AgentTmpl, k3OSNode)
	}

	misc.PanicOnError(err, "failed to create server install")

	cmdOperatorFactory := &pkg.CmdOperatorFactory{}
	if task.DryRun {
		cmdOperatorFactory.Create = ssh.NewDryRunCmdOperator
	} else {
		cmdOperatorFactory.Create = ssh.NewCmdOperator
	}

	return &installer{
		task:            task,
		resourceDir:     resourceDir,
		config:          configYaml,
		target:          k3OSNode,
		operatorFactory: cmdOperatorFactory,
	}
}

type installer struct {
	task            *OSInstallTask
	resourceDir     string
	config          *[]byte
	target          *model.K3OSNode
	operatorFactory *pkg.CmdOperatorFactory
}

func (ins *installer) Install() error {
	sshConfig, sshAgentCloseHandler, err := ssh.NewClientConfigFor(&ins.target.Node)
	misc.PanicOnError(err, "failed to create ssh config")
	defer sshAgentCloseHandler()

	address := ins.target.Node.Address
	sshAddress := fmt.Sprintf("%s:%d", address, 22)

	scpClient := scp.NewClient(sshAddress, sshConfig)
	err = scpClient.Connect()
	misc.PanicOnError(err, fmt.Sprintf("scp client failed to connect to %s", address))

	imageFile, err := os.Open(ins.task.GetImageFilePath(ins.resourceDir, ins.target.GetArch()))
	misc.PanicOnError(err, "failed to open image file")
	defer imageFile.Close()
	stat, err := imageFile.Stat()
	misc.PanicOnError(err, "failed to get file info")

	err = scpClient.Copy(bufio.NewReader(imageFile), fmt.Sprintf("~/%s", ins.task.GetImageFilename(ins.target.GetArch())), "0655", stat.Size())
	misc.PanicOnError(err, "failed to copy image file")

	// It's strange but we need to close and open for each file
	_ = scpClient.Session.Close()
	err = scpClient.Connect()
	misc.PanicOnError(err, fmt.Sprintf("scp client failed to connect to %s", address))
	defer scpClient.Session.Close()

	err = scpClient.Copy(bytes.NewReader(*ins.config), fmt.Sprintf("~/%s", "config.yaml"), "0655", int64(len(*ins.config)))
	misc.PanicOnError(err, "failed to copy config file")

	ctx := &pkg.CmdOperatorCtx{
		Address:         sshAddress,
		SSHClientConfig: sshConfig,
		EnableStdOut:    false,
	}

	operator, err := ins.operatorFactory.Create(ctx)
	misc.PanicOnError(err, fmt.Sprintf("failed to connect to %s", ctx.Address))

	fn := ins.task.GetImageFilename(ins.target.GetArch())
	result, err := operator.Execute(fmt.Sprintf("sudo tar zxvf %s --strip-components=1 -C /", fn))
	if err2 := errors.Wrap(err, fmt.Sprintf("failed to extract %s, result:\n %v", fn, result)); err2 != nil {
		return err2
	}

	result, err = operator.Execute("sudo cp config.yaml /k3os/system/config.yaml")
	if err2 := errors.Wrap(err, fmt.Sprintf("failed to install config:\n %v", result)); err2 != nil {
		return err2
	}

	_, _ = operator.Execute("sudo sync && sudo reboot -f")

	return nil
}
