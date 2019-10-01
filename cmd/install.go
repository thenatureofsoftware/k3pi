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
package cmd

import (
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg"
	cmd2 "github.com/TheNatureOfSoftware/k3pi/pkg/cmd"
	"github.com/kubernetes-sigs/yaml"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"os"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fn := viper.GetString("filename")

		var bytes []byte
		var err error
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			fmt.Println("data is being piped to stdin")
			bytes, err = ioutil.ReadAll(os.Stdin)
		} else {
			if fn == "" {
				fmt.Println(fmt.Errorf("Error: must specify --filename|-f"))
				os.Exit(1)
			}
			bytes, err = ioutil.ReadFile(fn)
		}

		if err != nil {
			log.Fatalf("Error reading input file, %s", err)
		}

		nodes := &[]pkg.Node{}
		err = yaml.Unmarshal(bytes, nodes)
		if err != nil {
			log.Fatalf("Error parsing nodes from file, %s", err)
		}

		if len(*nodes) == 0 {
			fmt.Println("No nodes found in file.")
			return
		}

		server := (*nodes)[0].GetK3sTarget([]string{})
		agents := []pkg.K3sTarget{}
		if len(*nodes) > 1 {
			for _, v := range (*nodes)[1:] {
				agents = append(agents, *v.GetK3sTarget([]string{}))
			}
		}

		_ = cmd2.MakeInstaller(&cmd2.InstallTask{
			Server: server,
			Agents: &agents,
		})
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().Bool("dry-run", false, "If true will print the install commands but never run them")
	installCmd.Flags().StringP("filename", "f", "", "If true will print the install commands but never run them")
	installCmd.Flags().Lookup("filename").NoOptDefVal = ""
	_ = viper.BindPFlag("dry-run", installCmd.Flags().Lookup("dry-run"))
	_ = viper.BindPFlag("filename", installCmd.Flags().Lookup("filename"))

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
