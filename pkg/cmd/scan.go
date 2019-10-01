package cmd

import (
    "fmt"
    "github.com/TheNatureOfSoftware/k3pi/pkg"
    "github.com/TheNatureOfSoftware/k3pi/pkg/misc"
    ssh2 "github.com/TheNatureOfSoftware/k3pi/pkg/ssh"
    "log"
    "strings"
)

type ScanRequest struct {
    Cidr, HostnameSubString string
    SSHSettings *ssh2.Settings
    UserCredentials map[string]string
}

func ScanForRaspberries(request *ScanRequest, hostScanner misc.HostScanner, cmdOperatorFactory *pkg.CmdOperatorFactory) (*[]pkg.Node, error) {

    settings := request.SSHSettings

    alive, err := hostScanner.ScanForAliveHosts(request.Cidr)
    if err != nil {
        return nil, err
    }

    config, closeSSHAgent, err := ssh2.NewClientConfig(request.SSHSettings)
    if err != nil {
        log.Fatalf("failed to create ssh config: %d", err)
    }
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
            if checkIfHostnameMatch(request.HostnameSubString, ctx, cmdOperatorFactory) {
                raspberries = append(raspberries, pkg.Node{
                    Address: ip,
                    Arch: arch,
                    Auth: pkg.Auth{
                        Type:   "ssh-key",
                        User:   settings.User,
                        SSHKey: settings.GetKeyPath(),
                    },
                })
            }
        } else {
            for username, password := range request.UserCredentials {
                altConfig, _, _ := ssh2.PasswordClientConfig(username, password)
                altCtx := *ctx
                altCtx.SSHClientConfig = altConfig
                if b, arch := checkArch(&altCtx, cmdOperatorFactory); b {
                    if checkIfHostnameMatch(request.HostnameSubString, &altCtx, cmdOperatorFactory) {
                        raspberries = append(raspberries, pkg.Node{
                            Address: ip,
                            Arch: arch,
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

    return  &raspberries, nil
}

func checkIfHostnameMatch(hostnameSubStr string, ctx *pkg.CmdOperatorCtx, cmdOperatorFactory *pkg.CmdOperatorFactory) bool {
    if len(hostnameSubStr) == 0 {
        return true
    }
    cmdOperator, err := cmdOperatorFactory.Create(ctx)
    if err != nil {
        return false
    }

    result, err := cmdOperator.Execute("hostname")
    if err != nil {
        return false
    }

    return strings.Contains(string(result.StdOut), hostnameSubStr)
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

    return true, strings.TrimSpace(string(result.StdOut))
}
