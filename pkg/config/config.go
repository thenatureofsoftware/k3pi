package config

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg"
	"github.com/kubernetes-sigs/yaml"
	"io/ioutil"
	"log"
	"text/template"
)

var ServerConfigTmpl = `hostname: {{.Node.Hostname}}
ssh_authorized_keys:
{{- range .SSHAuthorizedKeys}}
- "{{.}}"
{{- end}}
k3os:
  k3s_args:
  - server
  - "--disable-agent"
  - "--bind-address"
  - "{{.Node.Address}}"
  token: myclustersecret
  password: rancher
  dns_nameservers:
  - 8.8.8.8
  - 1.1.1.1
  ntp_servers:
  - 0.europe.pool.ntp.org
  - 1.europe.pool.ntp.org
  environment:
    INSTALL_K3S_VERSION: v0.9.1
`

var AgentConfigTmpl = `hostname: {{.Node.Hostname}}
ssh_authorized_keys:
{{- range .SSHAuthorizedKeys}}
- "{{.}}"
{{- end}}
k3os:
  k3s_args:
  - agent
  - "--node-ip"
  - "{{.Node.Address}}"
  server_url: https://{{.ServerIP}}:6443
  token: myclustersecret
  password: rancher
  dns_nameservers:
  - 8.8.8.8
  - 1.1.1.1
  ntp_servers:
  - 0.europe.pool.ntp.org
  - 1.europe.pool.ntp.org
  environment:
    INSTALL_K3S_VERSION: v0.9.1
`

type CloudConfig struct {
	Hostname          string   `json:"hostname"`
	SshAuthorizedKeys []string `json:"ssh_authorized_keys,omitempty"`
	K3os              K3os     `json:"k3os"`
}

type K3os struct {
	K3sArgs     []string          `json:"k3s_args,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
}

func (c *CloudConfig) LoadFromFile(filename string) *CloudConfig {
	yamlFile, _ := ioutil.ReadFile(filename)
	return c.LoadFromBytes(yamlFile)
}

func (c *CloudConfig) LoadFromBytes(content []byte) *CloudConfig {
	err := yaml.Unmarshal(content, c)
	if err != nil {
		log.Fatalf("%s", err)
	}
	return c
}

func NewServerConfig(configTmpl string, target *pkg.K3sTarget) (*[]byte, error) {
	tmpl := configTmpl
	if tmpl == "" {
		tmpl = ServerConfigTmpl
	}
	return generateConfig(tmpl, target)
}

func NewAgentConfig(configTmpl string, target *pkg.K3sTarget) (*[]byte, error) {
	tmpl := configTmpl
	if tmpl == "" {
		tmpl = AgentConfigTmpl
	}
	return generateConfig(tmpl, target)
}

func generateConfig(configTmpl string, target *pkg.K3sTarget) (*[]byte, error) {

	tmpl, err := template.New("cloud-config").Parse(configTmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cloud-config template: %d", err)
	}

	var b bytes.Buffer
	wr := bufio.NewWriter(&b)
	err = tmpl.Execute(wr, target)
	if err != nil {
		return nil, fmt.Errorf("failed to apply template to target %s: %v", target.Node.Address, err)
	}
	wr.Flush()
	configAsBytes := b.Bytes()
	return &configAsBytes, nil
}
