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
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
	"github.com/TheNatureOfSoftware/k3pi/pkg/misc"
	ssh2 "github.com/TheNatureOfSoftware/k3pi/pkg/ssh"
	"strings"
)

var SupportedArch = map[string]bool{
	"armv6l":  true,
	"armv7l":  true,
	"aarch64": true,
}

type ScanRequest struct {
	Cidr, HostnameSubString string
	SSHSettings             *ssh2.Settings
	UserCredentials         map[string]string
}

func ScanForRaspberries(request *ScanRequest, hostScanner misc.HostScanner, cmdOperatorFactory *pkg.CmdOperatorFactory) (*[]model.Node, error) {

	settings := request.SSHSettings

	alive, err := hostScanner.ScanForAliveHosts(request.Cidr)
	if err != nil {
		return nil, err
	}

	config, closeSSHAgent, err := ssh2.NewClientConfig(request.SSHSettings)
	misc.PanicOnError(err, "failed to create ssh config")
	defer closeSSHAgent()

	raspberries := []model.Node{}
	for i := range *alive {
		ip := (*alive)[i]
		address := fmt.Sprintf("%s:%s", ip, settings.Port)
		ctx := &pkg.CmdOperatorCtx{
			Address:         address,
			SSHClientConfig: config,
			EnableStdOut:    false,
		}

		if b, arch := checkArch(ctx, cmdOperatorFactory); b {
			if hn, ok := checkIfHostnameMatch(request.HostnameSubString, ctx, cmdOperatorFactory); ok {
				raspberries = append(raspberries, model.Node{
					Hostname: hn,
					Address:  ip,
					Arch:     arch,
					Auth: model.Auth{
						Type:   "ssh-key",
						User:   settings.User,
						SSHKey: settings.GetKeyPath(),
					},
				})
			}
		} else {
			for username, password := range request.UserCredentials {
				altConfig, _ := ssh2.PasswordClientConfig(username, password)
				altCtx := *ctx
				altCtx.SSHClientConfig = altConfig
				if b, arch := checkArch(&altCtx, cmdOperatorFactory); b {
					if hn, ok := checkIfHostnameMatch(request.HostnameSubString, &altCtx, cmdOperatorFactory); ok {
						raspberries = append(raspberries, model.Node{
							Hostname: hn,
							Address:  ip,
							Arch:     arch,
							Auth: model.Auth{
								Type:     "basic-auth",
								User:     username,
								Password: password,
							},
						})
						break
					}
				}
			}
		}
	}

	return &raspberries, nil
}

func checkIfHostnameMatch(hostnameSubStr string, ctx *pkg.CmdOperatorCtx, cmdOperatorFactory *pkg.CmdOperatorFactory) (string, bool) {

	cmdOperator, err := cmdOperatorFactory.Create(ctx)
	if err != nil {
		return "", false
	}

	result, err := cmdOperator.Execute("hostname")
	if err != nil {
		return "", false
	}

	hostname := strings.TrimSpace(string(result.StdOut))
	return hostname, strings.Contains(hostname, hostnameSubStr)
}

func checkArch(ctx *pkg.CmdOperatorCtx, cmdOperatorFactory *pkg.CmdOperatorFactory) (bool, string) {
	cmdOperator, err := cmdOperatorFactory.Create(ctx)
	if err != nil {
		return false, ""
	}

	result, err := cmdOperator.Execute("uname -m")
	if err != nil {
		return false, ""
	}

	arch := strings.TrimSpace(string(result.StdOut))
	if _, supported := SupportedArch[arch]; supported {
		return supported, arch
	} else {
		return false, ""
	}

}
