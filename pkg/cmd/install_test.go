package cmd

import (
    "github.com/TheNatureOfSoftware/k3pi/pkg/ssh"
    "testing"
)

func TestMakeInstaller(t *testing.T) {
    agents := []string{"192.168.1.11", "192.168.1.12", "192.168.1.13"}
    installers := MakeInstaller(true, &ssh.Settings{}, "192.168.1.10", agents)

    want := 4
    if count := len(*installers); count != want {
        t.Errorf("expected %d installers, got %d", want, count)
    }
}

