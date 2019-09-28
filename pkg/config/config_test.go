package config

import "testing"

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