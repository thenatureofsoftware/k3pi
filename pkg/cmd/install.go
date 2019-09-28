package cmd

import (
    "github.com/TheNatureOfSoftware/k3pi/pkg/ssh"
)

type Installer interface {
    Install() error
}

type InstallConfiguration struct {
    IP, Hostname string
    sshSettings *ssh.Settings
}

type installer struct {
    dryRun bool
    sshSettings *ssh.Settings
    master string
    nodes []string
}

func (ins installer) Install() error {
    panic("implement me!")
}

func MakeInstaller(dryRun bool, sshSettings *ssh.Settings, serverIpAddress string, agentIpAddresses []string) *[]Installer {
    installers := []Installer{}

    config := NewInstallConfiguration(serverIpAddress, "", sshSettings)
    installers = append(installers, MakeServerInstaller(dryRun, config))

    for _, agentIpAddress := range agentIpAddresses {
        config := NewInstallConfiguration(agentIpAddress, "", sshSettings)
        installers = append(installers, MakeAgentInstaller(dryRun, config))
    }

    return &installers
}

func NewInstallConfiguration(ipAddress string, hostname string, sshSettings *ssh.Settings) *InstallConfiguration {
    return &InstallConfiguration{
        IP:          ipAddress,
        Hostname:    hostname,
        sshSettings: sshSettings,
    }
}

func MakeServerInstaller(dryRun bool, config *InstallConfiguration) Installer {
    return &installer{}
}

func MakeAgentInstaller(dryRun bool, config *InstallConfiguration) Installer {
    return &installer{}
}
