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

// Package cmd include Cobra commands
package cmd

import (
	"fmt"
	pkgcmd "github.com/TheNatureOfSoftware/k3pi/pkg/cmd"
	"github.com/TheNatureOfSoftware/k3pi/pkg/install"
	"github.com/TheNatureOfSoftware/k3pi/pkg/misc"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
	"github.com/kubernetes-sigs/yaml"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"strings"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs k3os on selected nodes",
	Long: `Installs k3os on ARM devices, should be combined with the scan command.

	IMPORTANT! This will overwrite your existing installation.
	
	Examples:
	
	You should always run the install as a dry run first
	$ k3pi scan <scan args> | k3pi install --yes <install args> --dry-run
	
	Scan and install, confirm the install using --yes
	$ k3pi scan <scan args> | k3pi install --yes <install args>

	Installs k3os on all nodes in the file and selects <server ip> as server
	$ k3pi install --filename ./nodes.yaml --server <server ip>

	$ Installs k3os on all nodes as agents joining an existing server (server is not in nodes file)
	k3pi install --filename ./nodes.yaml -t <token|secret> --server <server ip>
`,
	Run: func(cmd *cobra.Command, args []string) {
		fn := viper.GetString(ParamFilename)

		var bytes []byte
		var err error

		if misc.DataPipedIn() {
			bytes, err = ioutil.ReadAll(os.Stdin)
		} else {
			if fn == "" {
				misc.ErrorExitWithMessage("must specify --filename|-f")
			}
			bytes, err = ioutil.ReadFile(fn)
		}
		misc.PanicOnError(err, "error reading input file")

		var nodes []*model.Node
		err = yaml.Unmarshal(bytes, &nodes)
		misc.ExitOnError(err, "error parsing nodes from file")

		if len(nodes) == 0 {
			misc.ErrorExitWithMessage("no nodes found in file")
		}

		k3OSVersion := viper.GetString(ParamK3OSVersionBindKey)
		if len(k3OSVersion) == 0 {
			misc.ErrorExitWithMessage("k3OS version ( -v|--version ) is empty")
		}
		sshKeys := viper.GetStringSlice(ParamSSHKeyInstallBindKey)
		server := viper.GetString(ParamServer)
		token := viper.GetString(ParamToken)
		dryRun := viper.GetBool(ParamInstallDryRunBindKey)
		hostnameSpec := &install.HostnameSpec{
			Pattern: viper.GetString(ParamHostnamePattern),
			Prefix:  viper.GetString(ParamHostnamePrefix),
		}

		serverConfigTmpl := loadTemplateFile(viper.GetString(ParamServerConfigTmpl))
		agentConfigTmpl := loadTemplateFile(viper.GetString(ParamAgentConfigTmpl))

		if len(sshKeys) == 0 {

			misc.ErrorExitWithMessage("at least one ssh key is required")

		} else if len(sshKeys) == 1 && sshKeys[0] == pkgcmd.K3OSDefaultSSHAuthorizedKey {

			idRsaPubFile, err := homedir.Expand(pkgcmd.K3OSDefaultSSHAuthorizedKey)
			msg := fmt.Sprintf("failed to read default ssh public key: %s", pkgcmd.K3OSDefaultSSHAuthorizedKey)
			misc.ExitOnError(err, msg)

			f, err := os.Open(idRsaPubFile)
			misc.ExitOnError(err, msg)
			defer f.Close()

			b, err := ioutil.ReadAll(f)
			misc.ExitOnError(err, msg)

			key := strings.Split(strings.TrimSpace(string(b)), " ")
			sshKeys = []string{fmt.Sprintf("%s %s", key[0], key[1])}
		}

		installArgs := &pkgcmd.InstallArgs{
			Nodes:        nodes,
			SSHKeys:      sshKeys,
			Token:        token,
			ServerID:     server,
			HostnameSpec: hostnameSpec,
			DryRun:       dryRun,
			Confirmed:    viper.GetBool(ParamConfirmInstall),
			Templates: &install.ConfigTemplates{
				ServerTmpl: serverConfigTmpl,
				AgentTmpl:  agentConfigTmpl,
			},
			K3OSVersion: k3OSVersion,
		}
		err = pkgcmd.Install(installArgs)
		misc.ExitOnError(err)
	},
}

func loadTemplateFile(configTmplFn string) string {
	if len(configTmplFn) != 0 {
		b, err := ioutil.ReadFile(configTmplFn)
		misc.PanicOnError(err, fmt.Sprintf("error reading template file: %s", configTmplFn))
		return string(b)
	}

	return ""
}

func init() {
	rootCmd.AddCommand(installCmd)

	installCmd.Flags().BoolP(ParamConfirmInstall, "y", false, "confirm the installation")
	installCmd.Flags().Bool(ParamDryRun, false, "if true will run the install but not execute commands")
	installCmd.Flags().String(ParamHostnamePattern, "%s%d", "hostname pattern, printf with %s and %d")
	installCmd.Flags().String(ParamHostnamePrefix, "k3s-node", "hostname prefix, (hostname = '<prefix><index>')")
	installCmd.Flags().StringP(ParamFilename, "f", "", "scan output file with all nodes")
	installCmd.Flags().StringP(ParamServer, "s", "", "ip address or hostname of the server node")
	installCmd.Flags().StringP(ParamToken, "t", "", "token or cluster secret for joining a server")
	installCmd.Flags().Lookup(ParamFilename).NoOptDefVal = ""
	installCmd.Flags().String(ParamServerConfigTmpl, "", "server k3OS config.yaml template file")
	installCmd.Flags().String(ParamAgentConfigTmpl, "", "agent k3OS config.yaml template file")
	installCmd.Flags().String(ParamVersion, model.DefaultK3OSVersion, fmt.Sprintf("k3OS version, default is %s", model.DefaultK3OSVersion))

	installCmd.Flags().StringSliceP(ParamSSHKey, "k", []string{pkgcmd.K3OSDefaultSSHAuthorizedKey}, "ssh authorized key that should be added to the rancher user")
	_ = viper.BindPFlag(ParamInstallDryRunBindKey, installCmd.Flags().Lookup(ParamDryRun))
	_ = viper.BindPFlag(ParamConfirmInstall, installCmd.Flags().Lookup(ParamConfirmInstall))
	_ = viper.BindPFlag(ParamFilename, installCmd.Flags().Lookup(ParamFilename))
	_ = viper.BindPFlag(ParamServer, installCmd.Flags().Lookup(ParamServer))
	_ = viper.BindPFlag(ParamSSHKeyInstallBindKey, installCmd.Flags().Lookup(ParamSSHKey))
	_ = viper.BindPFlag(ParamToken, installCmd.Flags().Lookup(ParamToken))
	_ = viper.BindPFlag(ParamHostnamePattern, installCmd.Flags().Lookup(ParamHostnamePattern))
	_ = viper.BindPFlag(ParamHostnamePrefix, installCmd.Flags().Lookup(ParamHostnamePrefix))
	_ = viper.BindPFlag(ParamServerConfigTmpl, installCmd.Flags().Lookup(ParamServerConfigTmpl))
	_ = viper.BindPFlag(ParamAgentConfigTmpl, installCmd.Flags().Lookup(ParamAgentConfigTmpl))
	_ = viper.BindPFlag(ParamK3OSVersionBindKey, installCmd.Flags().Lookup(ParamVersion))
}
