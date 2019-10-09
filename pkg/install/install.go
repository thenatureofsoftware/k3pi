package install

import (
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
	"github.com/TheNatureOfSoftware/k3pi/pkg/misc"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
)

type K3sUpgradeTask struct {
	model.Task
	Version string
	Nodes   model.Nodes
}

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
		_, err := os.Stat(resourceDir + pkg.PathSeparator + remoteAsset.Filename)
		if err == os.ErrNotExist {
			err := misc.DownloadAndVerify(remoteAsset)
			misc.PanicOnError(err, "failed to create resource directory")
		}
	}

	return resourceDir
}
