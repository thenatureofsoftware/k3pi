package install

import (
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
	"github.com/TheNatureOfSoftware/k3pi/pkg/ssh"
	"testing"
	"time"
)

type f1 struct {
}

func (f *f1) Supports(task interface{}) bool {
	return fmt.Sprintf("%T", task) == "string"
}

func (f *f1) MakeInstallers(task interface{}, resourceDir string) model.Installers {
	return model.Installers{}
}

func TestInstallerFactories_GetFactory(t *testing.T) {
	factory := f1{}
	factories := model.InstallerFactories{}
	factories = append(factories, &factory)

	if factories.GetFactory("") == nil {
		t.Fatal("factory not found")
	}

	if factories.GetFactory(1) != nil {
		t.Fatal("unknown factory")
	}
}

func TestWaitForNode(t *testing.T) {
	t.Skip("manual test")
	node := &model.Node{
		Address: model.NewAddress("192.168.1.111", 22),
	}
	sshSettings := &ssh.Settings{
		User:    "pirate",
		KeyPath: "~/.ssh/id_rsa",
		Port:    "22",
	}
	err := WaitForNode(node, sshSettings, time.Second*10)
	if err != nil {
		t.Error(err)
	}
}
