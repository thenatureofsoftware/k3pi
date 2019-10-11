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
	"github.com/TheNatureOfSoftware/k3pi/pkg/install"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
)

// UpgradeK3s upgrades k3s on a set of k3OS nodes
func UpgradeK3s(version string, nodes model.Nodes, dryRun bool) error {

	task := &install.K3sUpgradeTask{
		Task:          model.Task{DryRun: dryRun},
		Version:       version,
		Nodes:         nodes,
		ClientFactory: clientFactory,
	}

	factory := installerFactories.GetFactory(task)
	if factory == nil {
		fmt.Errorf("no installer factory found for task type: %T", task)
	}

	resourceDir := install.MakeResourceDir(task)

	installers := factory.MakeInstallers(task, resourceDir)

	return install.Run(installers)
}
