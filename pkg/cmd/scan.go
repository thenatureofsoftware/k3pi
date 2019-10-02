package cmd

import (
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg"
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

func ScanForRaspberries(request *ScanRequest, hostScanner misc.HostScanner, cmdOperatorFactory *pkg.CmdOperatorFactory) (*[]pkg.Node, error) {

	settings := request.SSHSettings

	alive, err := hostScanner.ScanForAliveHosts(request.Cidr)
	if err != nil {
		return nil, err
	}

	config, closeSSHAgent := ssh2.NewClientConfig(request.SSHSettings)
	defer closeSSHAgent()

	raspberries := []pkg.Node{}
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
				raspberries = append(raspberries, pkg.Node{
					Hostname: hn,
					Address:  ip,
					Arch:     arch,
					Auth: pkg.Auth{
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
						raspberries = append(raspberries, pkg.Node{
							Hostname: hn,
							Address:  ip,
							Arch:     arch,
							Auth: pkg.Auth{
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
