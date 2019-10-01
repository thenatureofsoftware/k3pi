package misc

import (
	"testing"
)

func TestHostScanner_ScanForAliveHosts_Localhost(t *testing.T) {
	scanner := NewHostScanner()

	alive, err := scanner.ScanForAliveHosts("127.0.0.1/32")
	if err != nil {
		t.Error(err)
	}

	verifyNumOfHosts(1, len(*alive), t)
}

func TestHostScanner_ScanForAliveHosts_Invalid_Cidr(t *testing.T) {
	scanner := NewHostScanner()

	alive, err := scanner.ScanForAliveHosts("I'm not a CIDR expr")
	if err != nil {
		t.Error(err)
	}

	verifyNumOfHosts(0, len(*alive), t)
}

func verifyNumOfHosts(want int, found int, t *testing.T) {
	if want != found {
		t.Errorf("wanted: %d but found: %d alive hosts", want, found)
	}
}
