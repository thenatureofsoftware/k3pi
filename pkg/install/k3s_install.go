package install

import (
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
)

type K3sInstallerFactory struct{}

func (k *K3sInstallerFactory) Supports(task interface{}) bool {
	tmpl := "%T"
	return fmt.Sprintf(tmpl, task) == fmt.Sprintf(tmpl, &K3sUpgradeTask{})
}

func (k *K3sInstallerFactory) MakeInstallers(task interface{}, resourceDir string) model.Installers {
	return model.Installers{}
}
