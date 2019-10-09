package cmd

import (
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg/install"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
)

func UpgradeK3s(version string, nodes model.Nodes, dryRun bool) error {

	task := &install.K3sUpgradeTask{
		Task:    model.Task{DryRun: dryRun},
		Version: version,
		Nodes:   nodes,
	}

	factory := installerFactories.GetFactory(task)
	if factory == nil {
		fmt.Errorf("no installer factory found for task type: %T", task)
	}

	return nil
}
