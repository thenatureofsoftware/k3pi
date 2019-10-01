package cmd

import (
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg"
	"github.com/TheNatureOfSoftware/k3pi/pkg/misc"
	"github.com/TheNatureOfSoftware/k3pi/pkg/ssh"
	"github.com/mitchellh/go-homedir"
	"io/ioutil"
	"log"
	"os"
)

type Installer interface {
	Install() error
}

type InstallTask struct {
	DryRun bool
	Server pkg.Node
	Agents []pkg.Node
}

type installer struct {
	dryRun      bool
	sshSettings *ssh.Settings
	master      string
	nodes       []string
}

func (ins installer) Install() error {
	panic("implement me!")
}

func MakeInstaller(task *InstallTask) *[]Installer {
	fmt.Printf("Installing %s as server and %d agents\n", task.Server.Address, len(task.Agents))

	resourceDir, err := makeResourceDir(task)
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(resourceDir)

	installers := []Installer{}

	installers = append(installers, makeServerInstaller(task))

	for _, agent := range task.Agents {
		installers = append(installers, makeAgentInstaller(task, &agent))
	}

	return &installers
}

func makeResourceDir(task *InstallTask) (string, error) {
	homedir, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	resourceDir, err := ioutil.TempDir(homedir, ".k3pi-")
	if err != nil {
		return "", err
	}

	imageFileTemplate := "k3os-rootfs-%s.tar.gz"
	checkSumFileTemplate := "sha256sum-%s.txt"
	images := make(map[string]string)
	images[fmt.Sprintf(imageFileTemplate, task.Server.GetArch())] = fmt.Sprintf(checkSumFileTemplate, task.Server.GetArch())
	for _, agent := range task.Agents {
		images[fmt.Sprintf(imageFileTemplate, agent.GetArch())] = fmt.Sprintf(checkSumFileTemplate, agent.GetArch())
	}

	url := "https://github.com/rancher/k3os/releases/download/v0.3.0/%s"
	for imageFile, checkSumFile := range images {
		download := misc.FileDownload{
			Filename:         imageFile,
			CheckSumFilename: checkSumFile,
			Url:              fmt.Sprintf(url, imageFile),
			CheckSumUrl:      fmt.Sprintf(url, checkSumFile),
		}
		err := misc.DownloadAndVerify(download)
		if err != nil {
			return resourceDir, err
		}
	}

	return resourceDir, nil
}

func makeServerInstaller(task *InstallTask) Installer {
	return &installer{}
}

func makeAgentInstaller(task *InstallTask, node *pkg.Node) Installer {
	return &installer{}
}
