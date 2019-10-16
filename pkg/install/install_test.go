package install

import (
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg/client"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
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
	factories := model.InstallerFactoriesT{}
	factories = append(factories, &factory)

	if factories.GetFactory("") == nil {
		t.Fatal("factory not found")
	}

	if factories.GetFactory(1) != nil {
		t.Fatal("unknown factory")
	}
}

func TestWaitForNode(t *testing.T) {
	node := &model.Node{
		Auth: model.Auth{
			Type:   model.AuthTypeSSHKey,
			User:   "rancher",
			SSHKey: "~/.ssh/id_rsa",
		},
		Address: model.NewAddress("192.168.1.111", 22),
	}
	cf, _ := client.NewFakeClientFactory()
	err := WaitForNode(cf, node, time.Second*10)
	if err != nil {
		t.Error(err)
	}
}
