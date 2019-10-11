package install

import (
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg/client"
	"github.com/TheNatureOfSoftware/k3pi/pkg/misc"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
	"github.com/pkg/errors"
	"strings"
)

const (
	// k3s release URL template
	K3sReleaseURLTmpl = "https://github.com/rancher/k3s/releases/download/%s/%s"
	// k3s binary filename template
	K3sBinFilenameTmpl = "k3s-%s"
	// k3s binary check sum filename template
	K3sBinCheckSumFilenameTmpl = "sha256sum-%s.txt"
)

// K3sUpgradeTask a task for upgrading k3s on a set of k3OS nodes
type K3sUpgradeTask struct {
	model.Task
	Version       string
	Nodes         model.Nodes
	ClientFactory client.Factory
}

// GetRemoteAssets gets all remote assets for a given k3s upgrade task
func (task *K3sUpgradeTask) GetRemoteAssets() model.RemoteAssets {

	var remoteAssets model.RemoteAssets

	for _, node := range task.Nodes {
		fn := fmt.Sprintf(K3sBinFilenameTmpl, node.GetArch("arm:armhf"))
		csfn := fmt.Sprintf(K3sBinCheckSumFilenameTmpl, node.GetArch())
		remoteAssets = append(remoteAssets, &model.RemoteAsset{
			Filename:         fn,
			FileUrl:          fmt.Sprintf(K3sReleaseURLTmpl, task.Version, fn),
			CheckSumFilename: csfn,
			CheckSumUrl:      fmt.Sprintf(K3sReleaseURLTmpl, task.Version, csfn),
		})
	}

	return remoteAssets
}

// K3sInstallerFactory factory for creating k3s upgrade installers
type K3sInstallerFactory struct{}

// Supports returns true if a given k3s installer factory supports the given task
func (k *K3sInstallerFactory) Supports(task interface{}) bool {
	tmpl := "%T"
	return fmt.Sprintf(tmpl, task) == fmt.Sprintf(tmpl, &K3sUpgradeTask{})
}

// MakeInstallers makes a set of installers for upgrading k3s on a set of k3OS nodes
func (k *K3sInstallerFactory) MakeInstallers(task interface{}, resourceDir string) model.Installers {
	upgradeTask, ok := task.(*K3sUpgradeTask)
	if !ok {
		misc.PanicOnError(fmt.Errorf("failed to cast to upgrade task, type was %T", task), "failed to make installers")
	}

	installers := model.Installers{}

	for _, node := range upgradeTask.Nodes {
		upgradeInstaller := makeK3sUpgradeInstaller(upgradeTask, resourceDir, node)
		installers = append(installers, upgradeInstaller)
	}

	return installers
}

func makeK3sUpgradeInstaller(task *K3sUpgradeTask, resourceDir string, node *model.Node) model.Installer {
	installer := &k3sInstaller{
		node:        node,
		task:        task,
		resourceDir: resourceDir,
	}
	return installer
}

type k3sInstaller struct {
	node        *model.Node
	task        *K3sUpgradeTask
	resourceDir string
}

func (ins *k3sInstaller) Install() error {
	node := ins.node

	nodeClient, err := ins.task.ClientFactory.Create(&node.Auth, &node.Address)
	if err != nil {
		return err
	}

	// copy file
	k3sBinFilename := fmt.Sprintf(K3sBinFilenameTmpl, node.GetArch("arm:armhf"))
	k3sBinFilenamePath := ins.resourceDir + PathSeparatorStr + k3sBinFilename
	nodeClient.Copy(k3sBinFilenamePath, "~/k3s")

	script := nodeClient.Cmd("sudo mount -o remount rw /k3os/system")
	script = script.Cmdf("sudo mkdir -p /k3os/system/k3s/%s", ins.task.Version)
	script = script.Cmdf("sudo cp ~/k3s /k3os/system/k3s/%s/", ins.task.Version)
	script = script.Cmdf("sudo chmod a+x /k3os/system/k3s/%s/k3s", ins.task.Version)
	script = script.Cmd("sudo /etc/init.d/k3s-service stop")
	script = script.Cmdf("sudo ln -sfn /k3os/system/k3s/%s /k3os/system/k3s/current", ins.task.Version)
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
