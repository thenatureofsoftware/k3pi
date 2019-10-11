package install

import (
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg/misc"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
	"github.com/TheNatureOfSoftware/k3pi/pkg/ssh"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"time"
)

const (
	// os path separator as string (for string concat)
	PathSeparatorStr = string(os.PathSeparator)
)

type installResult struct {
	installer model.Installer
	err       error
}

// Runs all installers in parallel
func Run(installers model.Installers) error {

	concurrentInstallers := 5
	installerCount := len(installers)
	if installerCount < concurrentInstallers {
		concurrentInstallers = installerCount
	}
	installChan := make(chan model.Installer, concurrentInstallers)
	doneChan := make(chan installResult, installerCount)

	fmt.Printf("Running install with %d concurrent installers\n", concurrentInstallers)

	for i := 0; i < concurrentInstallers; i++ {
		go func(installChan <-chan model.Installer, num int) {
			for installer := range installChan {
				fmt.Printf("Installer %d running ...\n", num)
				err := installer.Install()
				if err != nil {
					fmt.Printf("Installer %d running ... Failed\n", num)
				} else {
					fmt.Printf("Installer %d running ... OK\n", num)
				}
				doneChan <- installResult{
					installer: installer,
					err:       err,
				}
			}
		}(installChan, i)
	}

	for _, installer := range installers {
		installChan <- installer
	}

	var installErrors []error
	for i := 0; i < installerCount; i++ {
		result := <-doneChan
		if result.err != nil {
			installErrors = append(installErrors, errors.Wrap(result.err, fmt.Sprintf("install failed for install: %v", result.installer)))
		}
	}

	if len(installErrors) > 0 {
		//fmt.Printf("\r%s", strings.Repeat(" ", 35))
		fmt.Println("Install failed with errors")
		return fmt.Errorf("install errors: %s", installErrors)
	} else {
		fmt.Println("Install OK")
		return nil
	}
}

func MakeResourceDir(assetOwner model.RemoteAssetOwner) string {
	home, err := homedir.Dir()
	misc.PanicOnError(err, "failed to resolve home directory")

	resourceDir, err := ioutil.TempDir(home, ".k3pi-")
	misc.PanicOnError(err, "failed to create resource directory")

	for _, remoteAsset := range assetOwner.GetRemoteAssets() {
		_, err := os.Stat(resourceDir + PathSeparatorStr + remoteAsset.Filename)
		if os.IsNotExist(err) {
			err := misc.DownloadAndVerify(resourceDir, remoteAsset)
			misc.PanicOnError(err, "failed to create resource directory")
		}
	}

	return resourceDir
}

func WaitForNode(node *model.Node, sshSettings *ssh.Settings, timeout time.Duration) error {

	resolvedSSHSettings := resolveSSHSettings(sshSettings)

	clientConfig, sshAgentCloseHandler, err := ssh.NewClientConfig(resolveSSHSettings(sshSettings))
	misc.PanicOnError(err, "failed to create ssh agent")
	defer sshAgentCloseHandler()

	ctx := &ssh.CmdOperatorCtx{
		Address:         model.NewAddressStr(node.Address.IP, resolvedSSHSettings.Port),
		SSHClientConfig: clientConfig,
		EnableStdOut:    false,
	}

	timeToStop := time.Now().Add(timeout)
	for {
		_, err := ssh.NewCmdOperator(ctx)
		if err == nil {
			break
		} else if time.Now().After(timeToStop) {
			return fmt.Errorf("timeout waiting for node: %s", node.Address)
		}
		time.Sleep(time.Second * 2)
	}

	return nil
}

func resolveSSHSettings(sshSettings *ssh.Settings) *ssh.Settings {
	if sshSettings != nil {
		return sshSettings
	}
	return &ssh.Settings{
		User:    "rancher",
		KeyPath: "~/.ssh/id_rsa",
		Port:    "22",
	}
}
