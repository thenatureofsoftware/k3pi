package install

import (
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
	"testing"
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
