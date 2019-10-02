package config

import (
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg"
	"github.com/TheNatureOfSoftware/k3pi/pkg/misc"
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
	acctualSize := len(cloudConfig.SshAuthorizedKeys)
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
		SshAuthorizedKeys: []string{"github:foobar"},
		K3os: K3os{
			K3sArgs: []string{
				"server",
				"--disable-agent",
				"--bind-address",
				"127.0.0.1",
			},
		},
	}

	// Generate
	node := &pkg.Node{
		Hostname: "k3s-server",
		Address:  "127.0.0.1",
		Auth:     pkg.Auth{},
		Arch:     "aarch64",
	}
	configAsBytes, err := NewServerConfig("", &pkg.K3sTarget{
		SSHAuthorizedKeys: []string{"github:foobar"},
		Node:              node,
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
address: 127.0.0.1
arch: armv7l
auth:
  password: secret
  type: basic-auth
  user: root
`

	node := &pkg.Node{}
	err := yaml.Unmarshal([]byte(nodeYaml), node)
	misc.CheckError(err, "failed to unmarshal node")
	serverIp := "127.0.0.2"
	configAsBytes, err := NewAgentConfig("", &pkg.K3sTarget{
		SSHAuthorizedKeys: []string{"github:foobar"},
		Node:              node,
		ServerIP:          serverIp,
	})
	misc.CheckError(err, "failed to create agent config")
	fmt.Println(string(*configAsBytes))
}

func marshalToString(o interface{}) string {
	bytes, _ := yaml.Marshal(o)
	return string(bytes)
}
