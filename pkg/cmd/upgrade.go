/*
Copyright © 2019 The Nature of Software Nordic AB <lars@thenatureofsoftware.se>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

// Package cmd top package for handling Use Cases
package cmd

import (
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg/client"
	"github.com/TheNatureOfSoftware/k3pi/pkg/install"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
)

// Component k3OS component (OS, k3s, ...)
type Component int

const (
	// COS the OS component
	COS Component = iota
	// CK3s the k3s component
	CK3s
)

// Upgrade upgrades component on a set of nodes
func Upgrade(clientFactory *client.Factory, component Component, version string, nodes model.Nodes, dryRun bool) error {
	switch component {
	case COS:
		return upgradeK3OS(clientFactory, version, nodes, dryRun)
	case CK3s:
		return upgradeK3s(clientFactory, version, nodes, dryRun)
	default:
		return fmt.Errorf("unsupported component: %v", component)
	}
}

// upgradeK3s upgrades k3s on a set of k3OS nodes
func upgradeK3s(clientFactory *client.Factory, version string, nodes model.Nodes, dryRun bool) error {

	task := &install.K3sUpgradeTask{
		Task:          model.Task{DryRun: dryRun},
		Version:       version,
		Nodes:         nodes,
		ClientFactory: clientFactory,
	}

	factory := installerFactories.GetFactory(task)
	if factory == nil {
		return fmt.Errorf("no installer factory found for task type: %T", task)
	}

	resourceDir := install.MakeResourceDir(task)

	installers := factory.MakeInstallers(task, resourceDir)

	return install.Run(installers)
}

// upgradeK3OS upgrades k3OS on a set of k3OS nodes
func upgradeK3OS(clientFactory *client.Factory, version string, nodes model.Nodes, dryRun bool) error {

	task := &install.OSUpgradeTask{
		OSImageTask: install.OSImageTask{
			Task: model.Task{
				DryRun: dryRun,
			},
			Version:       version,
			ClientFactory: clientFactory,
		},
		Nodes: nodes,
	}

	factory := installerFactories.GetFactory(task)
	if factory == nil {
		return fmt.Errorf("no installer factory found for task type: %T", task)
	}

	resourceDir := install.MakeResourceDir(task)

	installers := factory.MakeInstallers(task, resourceDir)

	return install.Run(installers)
}
