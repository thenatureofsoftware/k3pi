package client

import (
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
	"strings"
	"testing"
)

func TestNewFakeClient(t *testing.T) {

	address := model.NewAddress("10.0.0.1", 22)
	cf, _ := NewFakeClientFactory(func(script *FakeScript) {
		script.Expect("whoami", "testuser")
	})

	client, err := cf.Create(&model.Auth{}, &address)

	if err != nil {
		t.Error(err)
	}

	output, err := client.Cmd("whoami").Output()
	if err != nil {
		t.Error(err)
	}

	s := strings.TrimSpace(string(output))
	if s != "testuser" {
		t.Error(fmt.Sprintf("cmd output don't match: %s", s))
	}
}
