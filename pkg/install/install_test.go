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
	clientFactor := client.Factory{Create: func(auth *model.Auth, address *model.Address) (i client.Client, e error) {
		return &client.FakeClient{}, nil
	}}
	node := &model.Node{
		Auth: model.Auth{
			Type:     model.AuthTypeSSHKey,
			User:     "rancher",
			SSHKey:   "~/.ssh/id_rsa",
		},
		Address: model.NewAddress("192.168.1.111", 22),
	}
	err := WaitForNode(clientFactor, node, time.Second*10)
	if err != nil {
		t.Error(err)
	}
}
