package client

import (
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
	"testing"
)

func TestNewFakeClient(t *testing.T) {

	client, err := NewFakeClient(&model.Auth{}, "10.0.0.1", 22)

	if err != nil {
		t.Error(err)
	}

	script := client.Cmd("whoami")
	fs := script.(*fakeScript)
	fs.output = "testuser"

	output, err := script.Output()
	if err != nil {
		t.Error(err)
	}

	s := string(output)
	if s != "testuser" {
		t.Error(fmt.Sprintf("cmd output don't match: %s", s))
	}
}
