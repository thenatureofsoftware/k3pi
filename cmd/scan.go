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
package cmd

import (
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg"
	cmd2 "github.com/TheNatureOfSoftware/k3pi/pkg/cmd"
	"github.com/TheNatureOfSoftware/k3pi/pkg/misc"
	"github.com/TheNatureOfSoftware/k3pi/pkg/ssh"
	"github.com/kubernetes-sigs/yaml"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strings"
)

// scanCmd represents the list command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scans the network for ARM devices",
	Long: `Scans the network for ARM devices with ssh enabled. The scan can use one SSH key
and multiple username and password combinations.

	Examples:

	# Scan using default SSH key in ~/.ssh/id_rsa, user root and CIDR 192.168.1.0/24
	$ k3pi scan

	# Scan using default SSH key in ~/.ssh/id_rsa and user foo
	$ k3pi scan --user foo --cidr 192.168.1.0/24

	# Scan using username and password
	$ k3pi scan --auth foo:bar --auth root:notsosecret

	# Scan filtering on hostname
	$ k3pi scan --substr pearl
`,
	Run: func(cmd *cobra.Command, args []string) {
		scanRequest := &cmd2.ScanRequest{
			Cidr:              viper.GetString(ParamCIDR),
			HostnameSubString: viper.GetString(ParamHostnameSubstring),
			SSHSettings:       sshSettings(),
			UserCredentials:   credentials(viper.GetStringSlice(ParamAuth)),
		}
		cmdOpFactory := &pkg.CmdOperatorFactory{Create: ssh.NewCmdOperator}
		nodes, err := cmd2.ScanForRaspberries(scanRequest, misc.NewHostScanner(), cmdOpFactory)
		misc.ExitOnError(err, "node scan failed")

		y, err := yaml.Marshal(nodes)
		misc.ExitOnError(err, "node scan failed")

		fmt.Print(string(y))
	},
}

// Splits slice of <username>:<password> and returns a map
func credentials(basicAuths []string) map[string]string {
	c := make(map[string]string)
	for _, v := range basicAuths {
		parts := strings.Split(v, ":")
		if len(parts) == 2 {
			c[parts[0]] = parts[1]
		}
	}
	return c
}

func init() {
	rootCmd.AddCommand(scanCmd)
	scanCmd.Flags().String(ParamUser, "root", "username for ssh login")
	scanCmd.Flags().String(ParamSSHKey, "~/.ssh/id_rsa", "ssh key to use for remote login")
	scanCmd.Flags().Int(ParamSSHPort, 22, "port on which to connect for ssh")
	scanCmd.Flags().String(ParamCIDR, "192.168.1.0/24", "CIDR to scan for members")
	scanCmd.Flags().String(ParamHostnameSubstring, "", "Substring that should be part of hostname")
	scanCmd.Flags().StringSliceP(ParamAuth, "a", []string{}, "Username and password separated with ':' for authentication")
	_ = viper.BindPFlag(ParamUser, scanCmd.Flags().Lookup(ParamUser))
	_ = viper.BindPFlag(ParamSSHKey, scanCmd.Flags().Lookup(ParamSSHKey))
	_ = viper.BindPFlag(ParamSSHPort, scanCmd.Flags().Lookup(ParamSSHPort))
	_ = viper.BindPFlag(ParamCIDR, scanCmd.Flags().Lookup(ParamCIDR))
	_ = viper.BindPFlag(ParamHostnameSubstring, scanCmd.Flags().Lookup(ParamHostnameSubstring))
	_ = viper.BindPFlag(ParamAuth, scanCmd.Flags().Lookup(ParamAuth))
}

func sshSettings() *ssh.Settings {
	return &ssh.Settings{
		KeyPath: viper.GetString(ParamSSHKey),
		User:    viper.GetString(ParamUser),
		Port:    viper.GetString(ParamSSHPort)}
}
