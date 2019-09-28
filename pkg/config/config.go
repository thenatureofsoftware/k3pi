package config

import (
    "gopkg.in/yaml.v2"
    "io/ioutil"
)

type CloudConfig struct {
    Hostname string `yaml:"hostname"`
    SshAuthorizedKeys []string `yaml:"ssh_authorized_keys"`
    K3os K3os `yaml:"k3os"`
}

type K3os struct {
    K3sArgs []string `yaml:"k3s_args"`
    Environment map[string]string
}

func (c *CloudConfig) LoadFromFile(filename string) *CloudConfig {
    yamlFile, _ := ioutil.ReadFile(filename)
    return c.LoadFromBytes(yamlFile)
}

func (c *CloudConfig) LoadFromBytes(content []byte) *CloudConfig {
    _ = yaml.Unmarshal(content, c)
    return c
}
