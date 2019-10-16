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
	"github.com/TheNatureOfSoftware/k3pi/pkg/client"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	host1 = "10.0.0.1"
	host2 = "10.0.0.2"
)

var fakeClient *client.FakeClient
var hostScannerResult = []string{host1, host2}

type mockHostScanner struct {
	returnError bool
}

func (s mockHostScanner) ScanForAliveHosts(cidr string) (*[]string, error) {
	if s.returnError {
		return nil, fmt.Errorf("failed to scan for hosts with CIDR: %s", cidr)
	}
	return &hostScannerResult, nil
}

func createScanRequest() *ScanRequest {
	scanRequest := &ScanRequest{
		Cidr:              "127.0.0.1/32",
		HostnameSubString: "",
		SSHAuth: &model.Auth{
			Type:   model.AuthTypeSSHKey,
			User:   "",
			SSHKey: "~/.ssh/id_rsa",
		},
		Port:            22,
		UserCredentials: make(map[string]string),
	}
	return scanRequest
}

func TestScanForNodes(t *testing.T) {
	clientFactory, _ := client.NewFakeClientFactory(func(script *client.FakeScript) {
		script.Expect("uname -m", "aarch64")
		script.Expect("hostname", "host1")
		script.Expect("uname -m", "armv7l")
		script.Expect("hostname", "host2")
	})

	request := createScanRequest()
	nodes, err := ScanForNodes(clientFactory, request, &mockHostScanner{})

	assert.NoError(t, err)
	assert.Len(t, *nodes, 2)
	assert.Equal(t, "host1", (*nodes)[0].Hostname)
	assert.Equal(t, "host2", (*nodes)[1].Hostname)
}

func TestScanForNodes_FilterOnHostname(t *testing.T) {
	clientFactory, _ := client.NewFakeClientFactory(func(script *client.FakeScript) {
		script.Expect("uname -m", "aarch64")
		script.Expect("hostname", "host1")
		script.Expect("uname -m", "armv7l")
		script.Expect("hostname", "host2")
	})
	request := createScanRequest()
	request.HostnameSubString = "2"
	nodes, err := ScanForNodes(clientFactory, request, &mockHostScanner{})

	assert.NoError(t, err)
	assert.Len(t, *nodes, 1)
	assert.Equal(t, "host2", (*nodes)[0].Hostname)
}

func TestScanRequest_GetAuths(t *testing.T) {
	cred := make(map[string]string)
	username := "test1"
	password := "mysecret"
	cred[username] = password
	req := &ScanRequest{
		Cidr:              "",
		HostnameSubString: "",
		SSHAuth: &model.Auth{
			Type:   model.AuthTypeSSHKey,
			User:   username,
			SSHKey: "~/.ssh/id_rsa",
		},
		Port:            22,
		UserCredentials: cred,
	}

	auths := req.GetAuths()

	assert.Len(t, auths, 2)
	for i := range auths {
		assert.Equal(t, auths[i].User, username)
	}
	assert.Contains(t, auths[0].SSHKey, "id_rsa")
	assert.Equal(t, auths[1].Password, password)
}

