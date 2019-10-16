/*
Copyright © 2019 Lars Mogren <lars@thenatureofsoftware.se>

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

// Package cmd include Cobra commands
package cmd

import (
	"github.com/TheNatureOfSoftware/k3pi/pkg/client"
	pkgcmd "github.com/TheNatureOfSoftware/k3pi/pkg/cmd"
	"github.com/TheNatureOfSoftware/k3pi/pkg/misc"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
	"github.com/kubernetes-sigs/yaml"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var components = map[string]pkgcmd.Component{}

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrades k3s or k3os on all nodes.",
	Long: `Upgrades all nodes to the specified version of k3os or k3s.
	Example:
	
	Upgrades k3s on all nodes from a nodes file
	$ k3pi upgrade -f ./nodes.yaml -c k3s --version <k3s version>

	Upgrades k3os on all nodes from a nodes file
	$ k3pi upgrade -f ./nodes.yaml -c os --version <k3os version>
`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var bytes []byte

		dryRun := viper.GetBool(ParamUpgradeDryRunBindKey)
		fn := viper.GetString(ParamUpgradeFilename)
		targetVersion := viper.GetString(ParamK3sVersionBindKey)
		cstr := strings.ToLower(viper.GetString(ParamComponent))

		var component pkgcmd.Component
		if c, ok := components[cstr]; ok {
			component = c
		} else {
			misc.ErrorExitWithMessage("unsupported component")
		}

		if misc.DataPipedIn() {
			bytes, err = ioutil.ReadAll(os.Stdin)
		} else {
			if fn == "" {
				misc.ErrorExitWithMessage("must specify --filename|-f")
			}
			bytes, err = ioutil.ReadFile(fn)
		}
		misc.PanicOnError(err, "error reading input file")

		var nodes model.Nodes
		err = yaml.Unmarshal(bytes, &nodes)
		misc.ExitOnError(err, "error parsing nodes from file")

		if len(nodes) == 0 {
			misc.ErrorExitWithMessage("no nodes found in file")
		}

		if len(targetVersion) == 0 {
			misc.ErrorExitWithMessage("invalid or missing k3s target version ( --version|-v )")
		}

		err = pkgcmd.Upgrade(client.NewClientFactory(), component, targetVersion, nodes, dryRun)
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)

	components["os"] = pkgcmd.COS
	components["k3s"] = pkgcmd.CK3s

	upgradeCmd.Flags().Bool(ParamDryRun, false, "if true will run the install but not execute commands")
	upgradeCmd.Flags().StringP(ParamVersion, "v", "", "target k3s version")
	upgradeCmd.Flags().StringP(ParamFilename, "f", "", "scan output file with all nodes")
	upgradeCmd.Flags().StringP(ParamComponent, "c", "", "witch component to upgrade os|k3s")
	_ = upgradeCmd.MarkFlagRequired(ParamVersion)
	_ = upgradeCmd.MarkFlagRequired(ParamComponent)

	_ = viper.BindPFlag(ParamUpgradeDryRunBindKey, upgradeCmd.Flags().Lookup(ParamDryRun))
	_ = viper.BindPFlag(ParamK3sVersionBindKey, upgradeCmd.Flags().Lookup(ParamVersion))
	_ = viper.BindPFlag(ParamUpgradeFilename, upgradeCmd.Flags().Lookup(ParamFilename))
	_ = viper.BindPFlag(ParamComponent, upgradeCmd.Flags().Lookup(ParamComponent))
}
