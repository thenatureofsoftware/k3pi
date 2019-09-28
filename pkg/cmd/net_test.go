package cmd

import (
	"fmt"
    "github.com/TheNatureOfSoftware/k3pi/pkg/ssh"
    "testing"
)

func TestScanForPi(t *testing.T) {
    raspberries, err := ScanForRaspberries("192.168.1.111/32", "rpi5",
    	&ssh.Settings{KeyPath: "~/.ssh/id_rsa", User: "tnos", Port: "22"})

	if err != nil {
		t.Errorf("scan for Raspberry Pi:s failed: %d", err)
	}

	fmt.Println(raspberries)
}
