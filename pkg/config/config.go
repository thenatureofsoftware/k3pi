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
	"bufio"
	"bytes"
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
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
  - "--bind-address"
  - "{{.Node.Address.IP}}"
  token: {{.Token}}
  password: rancher
  dns_nameservers:
  - 8.8.8.8
  - 1.1.1.1
  ntp_servers:
  - 0.europe.pool.ntp.org
  - 1.europe.pool.ntp.org`

var AgentConfigTmpl = `hostname: {{.Node.Hostname}}
ssh_authorized_keys:
{{- range .SSHAuthorizedKeys}}
- "{{.}}"
{{- end}}
k3os:
  k3s_args:
  - agent
  - "--node-ip"
  - "{{.Node.Address.IP}}"
  server_url: https://{{.ServerIP}}:6443
  token: {{.Token}}
  password: rancher
  dns_nameservers:
  - 8.8.8.8
  - 1.1.1.1
  ntp_servers:
  - 0.europe.pool.ntp.org
  - 1.europe.pool.ntp.org`

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

func NewServerConfig(configTmpl string, target *model.K3OSNode) (*[]byte, error) {
	tmpl := configTmpl
	if tmpl == "" {
		tmpl = ServerConfigTmpl
	}
	return generateConfig(tmpl, target)
}

func NewAgentConfig(configTmpl string, target *model.K3OSNode) (*[]byte, error) {
	tmpl := configTmpl
	if tmpl == "" {
		tmpl = AgentConfigTmpl
	}
	return generateConfig(tmpl, target)
}

func generateConfig(configTmpl string, target *model.K3OSNode) (*[]byte, error) {

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
