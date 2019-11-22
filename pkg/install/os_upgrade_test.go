package install

import (
    "fmt"
    "github.com/TheNatureOfSoftware/k3pi/pkg/client"
    "github.com/TheNatureOfSoftware/k3pi/pkg/model"
    "github.com/TheNatureOfSoftware/k3pi/test"
    "github.com/stretchr/testify/assert"
    "testing"
)

func TestOSUpgradeInstallerFactory_Supports_Should_Support(t *testing.T) {
    supported := &OSUpgradeTask{}
    factory := OSUpgradeInstallerFactory{}
    assert.True(t, factory.Supports(supported))
}

func TestOSUpgradeInstallerFactory_Supports_Should_Not_Support(t *testing.T) {
    unsupported := &K3sUpgradeTask{}
    factory := OSUpgradeInstallerFactory{}
    assert.False(t, factory.Supports(unsupported))
}

func TestOSUpgradeInstallerFactory_MakeInstallers(t *testing.T) {
    clientFactory, _ := client.NewFakeClientFactory()

    nodes := test.CreateNodes()
    task := createOSUpgradeTask(clientFactory, nodes)

    count := len(nodes)
    f := OSUpgradeInstallerFactory{}
    installers := f.MakeInstallers(task, "")

    assert.Len(t, installers, count)
}

func TestOsUpgradeInstaller_Install(t *testing.T) {
    cf, script := client.NewFakeClientFactory()
    nodes := test.CreateNodes()
    node := nodes[0]

    task := createOSUpgradeTask(cf, nodes)
    installer := osUpgradeInstaller{
        task: task,
        node: node,
    }

    script.Expect(fmt.Sprintf("sudo tar zxf %s", task.GetImageFilename(node.GetArch())), "")

    err := installer.Install()

    assert.NoError(t, err, "install returned an error")
    assert.False(t, script.HasOutstandingCmds(), "there were expected commands not invoked %v", script.Interactions)
}

func TestOsUpgradeInstaller_Install_WithUpgradeRootfsScript(t *testing.T) {
    cf, script := client.NewFakeClientFactory()
    nodes := test.CreateNodes()
    node := nodes[0]
    task := createOSUpgradeTask(cf, nodes)

    installer := osUpgradeInstaller{
        task: task,
        node: node,
    }

    script.Expect("which k3os-upgrade-rootfs", "/usr/sbin/k3os-upgrade-rootfs")
    script.Expect(fmt.Sprintf("sudo K3OS_VERSION=%s k3os-upgrade-rootfs", task.Version), "Upgrade complete! Please reboot.")

    err := installer.Install()

    assert.NoError(t, err, "install returned an error")
    assert.False(t, script.HasOutstandingCmds(), "there were expected commands not invoked %v", script.Interactions)
}

func createOSUpgradeTask(clientFactory *client.Factory, nodes model.Nodes) *OSUpgradeTask {
    return &OSUpgradeTask{
        OSImageTask: OSImageTask{
            Task:          model.Task{},
            Version:       "v0.4.0",
            ClientFactory: clientFactory,
        },
        Nodes: nodes,
    }
}

