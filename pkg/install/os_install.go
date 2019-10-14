package install

import (
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg/client"
	"github.com/TheNatureOfSoftware/k3pi/pkg/config"
	"github.com/TheNatureOfSoftware/k3pi/pkg/misc"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
	"github.com/pkg/errors"
	"strings"
)

const (
	// K3OSImageFilenameTmpl OS image filename template
	K3OSImageFilenameTmpl = "k3os-rootfs-%s.tar.gz"
	// K3OSCheckSumFileTmpl OS image check sum filename template
	K3OSCheckSumFileTmpl = "sha256sum-%s.txt"
	// K3OSReleaseURLTmpl OS release url template
	K3OSReleaseURLTmpl = "https://github.com/rancher/k3os/releases/download/%s/%s"
)

// ConfigTemplates config.yaml go templates for generating server and agent config
type ConfigTemplates struct {
	ServerTmpl, AgentTmpl string
}

// HostnameSpec spec for generating hostnames (for nodes)
type HostnameSpec struct {
	Pattern, Prefix string
}

// GetHostname gets a generated hostname given the node list index
func (h *HostnameSpec) GetHostname(index int) string {
	return fmt.Sprintf(h.Pattern, h.Prefix, index)
}

// OSInstallTask task for installing k3OS
type OSInstallTask struct {
	model.Task
	Server        *model.K3OSNode
	Agents        model.K3OSNodes
	Version       string
	Templates     *ConfigTemplates
	ClientFactory client.Factory
}

// GetRemoteAssets gets all remote assets for all nodes in this task
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
			FileURL:          task.GetImageFileURL(arch),
			CheckSumFilename: task.GetImageCheckSumFilename(arch),
			CheckSumURL:      task.GetImageCheckSumURL(arch),
		})
	}

	return resources
}

// GetImageFilePath returns the full path of the image file given an architecture (arm, arm64)
func (task *OSInstallTask) GetImageFilePath(resourceDir string, arch string) string {
	return fmt.Sprintf("%s%s%s", resourceDir, PathSeparatorStr, task.GetImageFilename(arch))
}

// GetImageCheckSumFilePath returns the full path of the image check sum file given an architecture (arm, arm64)
func (task *OSInstallTask) GetImageCheckSumFilePath(resourceDir string, arch string) string {
	return fmt.Sprintf("%s%s%s", resourceDir, PathSeparatorStr, task.GetImageCheckSumFilename(arch))
}

// GetImageFilename returns image filename given an architecture (arm, arm64)
func (task *OSInstallTask) GetImageFilename(arch string) string {
	return fmt.Sprintf(K3OSImageFilenameTmpl, arch)
}

// GetImageCheckSumFilename returns image check sum filename given an architecture (arm, arm64)
func (task *OSInstallTask) GetImageCheckSumFilename(arch string) string {
	return fmt.Sprintf(K3OSCheckSumFileTmpl, arch)
}

// GetImageFileURL returns image file url given an architecture (arm, arm64)
func (task *OSInstallTask) GetImageFileURL(arch string) string {
	return fmt.Sprintf(K3OSReleaseURLTmpl, task.Version, task.GetImageFilename(arch))
}

// GetImageCheckSumURL returns image check sum file url given an architecture (arm, arm64)
func (task *OSInstallTask) GetImageCheckSumURL(arch string) string {
	return fmt.Sprintf(K3OSReleaseURLTmpl, task.Version, task.GetImageCheckSumFilename(arch))
}

// OSInstallerFactory factory for creating k3OS installers
type OSInstallerFactory struct{}

// Supports returns true if this factory supports creating an installer for the given task
func (o OSInstallerFactory) Supports(task interface{}) bool {
	return fmt.Sprintf("%T", task) == fmt.Sprintf("%T", &OSInstallTask{})
}

// MakeInstallers creates installers for the given task
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

	//cmdOperatorFactory := &ssh.CmdOperatorFactory{}
	//if task.DryRun {
	//	cmdOperatorFactory.Create = ssh.NewDryRunCmdOperator
	//} else {
	//	cmdOperatorFactory.Create = ssh.NewCmdOperator
	//}

	return &installer{
		task:            task,
		resourceDir:     resourceDir,
		config:          configYaml,
		target:          k3OSNode,
		//operatorFactory: cmdOperatorFactory,
	}
}

type installer struct {
	task            *OSInstallTask
	resourceDir     string
	config          *[]byte
	target          *model.K3OSNode
	//operatorFactory *ssh.CmdOperatorFactory
}

// Install installs k3OS
func (ins *installer) Install() error {

	sshClient, err := ins.task.ClientFactory.Create(&ins.target.Auth, &ins.target.Address)
	misc.PanicOnError(err, "failed to create SSH client")

	err = sshClient.Copy(ins.task.GetImageFilePath(ins.resourceDir, ins.target.GetArch()), fmt.Sprintf("~/%s", ins.task.GetImageFilename(ins.target.GetArch())))
	misc.PanicOnError(err, "failed to copy image file")

	err = sshClient.CopyBytes(ins.config, fmt.Sprintf("~/%s", "config.yaml"))
	misc.PanicOnError(err, "failed to copy config file")

	fn := ins.task.GetImageFilename(ins.target.GetArch())
	script := sshClient.Cmdf("sudo tar zxvf %s --strip-components=1 -C /", fn)
	script = script.Cmd("sudo cp config.yaml /k3os/system/config.yaml")
	script = script.Cmd("sudo sync")
	script = script.Cmd("sudo reboot -d 1 &")

	if ins.task.DryRun {
		return nil
	}

	out, err := script.Output()
	if err != nil {
		stdErr := strings.TrimSpace(string(out))
		fmt.Println(stdErr)
		return errors.Wrap(err, stdErr)
	}

	return nil
}
