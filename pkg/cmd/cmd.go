package cmd

import (
	"github.com/TheNatureOfSoftware/k3pi/pkg/install"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
)

var installerFactories = model.InstallerFactories{&install.OSInstallerFactory{}}
