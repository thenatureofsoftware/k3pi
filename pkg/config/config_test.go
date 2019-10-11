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
package config

import (
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg/misc"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
	"github.com/kubernetes-sigs/yaml"
	"testing"
)

var cloudConfigYaml = `
hostname: pi
ssh_authorized_keys:
- ssh-rsa AAAAB3NzaC1yc2EAAAADAQAB
- github:tnos
k3os:
  k3s_args:
  - server
  - "--disable-agent"
  environment:
    http_proxy: http://myserver
    http_proxys: http://myserver
`

var configTmpl = `
hostname: {{.Node.Hostname}}
ssh_authorized_keys:
- github:foobar
k3os:
  k3s_args:
  - server
  - "--disable-agent"
  - "--bind-address {{.Node.Address}}"
  environment:
    http_proxy: http://myserver
    http_proxys: http://myserver
`

func TestCloudConfig_LoadFrom(t *testing.T) {
	cloudConfig := &CloudConfig{}
	cloudConfig.LoadFromBytes([]byte(cloudConfigYaml))

	if cloudConfig.Hostname != "pi" {
		t.Fail()
	}

	expectedSize := 2
	acctualSize := len(cloudConfig.SSHAuthorizedKeys)
	if acctualSize != expectedSize {
		t.Errorf("expected %d keys, found %d", expectedSize, acctualSize)
	}

	expectedArgCount := 2
	acctualArgCount := len(cloudConfig.K3os.K3sArgs)
	if acctualArgCount != expectedArgCount {
		t.Errorf("expected %d k3s arguments, found %d", expectedArgCount, acctualArgCount)
	}

	expectedSize = 2
	acctualSize = len(cloudConfig.K3os.Environment)
	if acctualSize != expectedSize {
		t.Errorf("expected %d env variables, found %d", expectedSize, acctualSize)
	}
}

func TestNewServerConfig(t *testing.T) {
	// This is what we wan't
	want := CloudConfig{
		Hostname:          "k3s-server",
		SSHAuthorizedKeys: []string{"github:foobar"},
		K3os: K3os{
			K3sArgs: []string{
				"server",
				"--bind-address",
				"10.0.0.1",
			},
		},
	}

	// Generate
	node := &model.Node{
		Hostname: "k3s-server",
		Address:  model.ParseAddress("10.0.0.1:22"),
		Auth:     model.Auth{},
		Arch:     "aarch64",
	}
	configAsBytes, err := NewServerConfig("", &model.K3OSNode{
		SSHAuthorizedKeys: []string{"github:foobar"},
		Node:              *node,
	})

	if err != nil {
		t.Error(err)
	}

	// This is what we got
	actual := CloudConfig{}
	actual.LoadFromBytes(*configAsBytes)

	wantAsYaml := marshalToString(want)
	actualAsYaml := marshalToString(actual)
	if wantAsYaml != actualAsYaml {
		t.Errorf("wanted:\n%s\nactual:\n%s\n", wantAsYaml, actualAsYaml)
	}
}

func TestNewAgentConfig(t *testing.T) {
	var nodeYaml = `
hostname: test
address:
  ip: 127.0.0.1
  port: 22
arch: armv7l
auth:
  password: secret
  type: basic-auth
  user: root
`

	node := &model.Node{}
	err := yaml.Unmarshal([]byte(nodeYaml), node)
	misc.PanicOnError(err, "failed to unmarshal node")
	serverIP := "127.0.0.2"
	configAsBytes, err := NewAgentConfig("", &model.K3OSNode{
		SSHAuthorizedKeys: []string{"github:foobar"},
		Node:              *node,
		ServerIP:          serverIP,
	})
	misc.PanicOnError(err, "failed to create agent config")
	fmt.Println(string(*configAsBytes))
}

func marshalToString(o interface{}) string {
	bytes, _ := yaml.Marshal(o)
	return string(bytes)
}
