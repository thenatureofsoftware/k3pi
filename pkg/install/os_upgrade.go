package install

import (
    "fmt"
    "github.com/TheNatureOfSoftware/k3pi/pkg/client"
    "github.com/TheNatureOfSoftware/k3pi/pkg/misc"
    "github.com/TheNatureOfSoftware/k3pi/pkg/model"
    "github.com/pkg/errors"
    "reflect"
    "strings"
)

// OSUpgradeTask task for upgrading k3OS
type OSUpgradeTask struct {
	OSImageTask
	Nodes model.Nodes
}

// GetRemoteAssets gets all remote assets (k3OS image files) for all nodes in this task
func (t *OSUpgradeTask) GetRemoteAssets() model.RemoteAssets {
	return createRemoteAssets(t.OSImageTask, t.Nodes)
}

// OSUpgradeInstallerFactory installer factory for creating OS upgrade installer
type OSUpgradeInstallerFactory struct {}

// Supports returns true if the factory supports the given task
func (f *OSUpgradeInstallerFactory) Supports(task interface{}) bool {
    t := reflect.TypeOf(task)
    return t.AssignableTo(reflect.TypeOf(&OSUpgradeTask{}))
}

// MakeInstallers makes installers for the given task
func (f *OSUpgradeInstallerFactory) MakeInstallers(task interface{}, resourceDir string) model.Installers {

    osUpgradeTask := task.(*OSUpgradeTask)

    var installers model.Installers
    for _, node := range osUpgradeTask.Nodes {
        installers = append(installers, &osUpgradeInstaller{
            task: osUpgradeTask,
            node: node,
            resourceDir:resourceDir,
        })
    }

    return installers
}

type osUpgradeInstaller struct {
    task *OSUpgradeTask
    node *model.Node
    resourceDir string
}

func (ins *osUpgradeInstaller) Install() error {

    sshClient, err := ins.task.ClientFactory.Create(&ins.node.Auth, &ins.node.Address)
    if err != nil {
        return err
    }

    arch := ins.node.GetArch()
    fn := ins.task.GetImageFilename(arch)

    var script client.Script
    if hasUpgradeRootfs(sshClient) {
        script = sshClient.Cmdf("sudo k3os-upgrade-rootfs %s", ins.task.Version)
        script = script.Cmd("sudo reboot -d 1 &")
    } else {
        err = sshClient.Copy(ins.task.GetImageFilePath(ins.resourceDir, arch), fmt.Sprintf("~/%s", ins.task.GetImageFilename(arch)))
        misc.PanicOnError(err, "failed to copy image file")

        script = sshClient.Cmdf("sudo /etc/init.d/k3s-service stop")
        script = script.Cmdf("sudo mkdir -p /var/tnos/k3pi")
        script = script.Cmdf("sudo cp /k3os/system/config.yaml /var/tnos/k3pi/")
        script = script.Cmd("sudo mount -o remount rw /k3os/system")

        script = script.Cmdf("sudo tar zxf %s", fn)
        script = script.Cmdf("sudo cp -R %s/k3os/system/* /k3os/system", ins.task.Version)
        script = script.Cmdf("sudo rm -rf %s %s", fn, ins.task.Version)

        script = script.Cmd("sudo cp /var/tnos/k3pi/config.yaml /k3os/system/config.yaml")
        script = script.Cmd("sudo sync")
        script = script.Cmd("sudo reboot -d 1 &")
    }
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

func hasUpgradeRootfs(client client.Client) bool {
    path, err := client.Cmd("which k3os-upgrade-rootfs").Output()
    if err != nil {
        return false
    }
    return len(strings.TrimSpace(string(path))) > 0
}


