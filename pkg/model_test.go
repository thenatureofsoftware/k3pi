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
package pkg

import (
	"fmt"
	"github.com/kubernetes-sigs/yaml"
	"os"
	"testing"
)

var msg = "\nexpected: %v\nactual: %v"

var node = &Node{
	Arch: "aarch64",
}

func TestK3sTarget_GetImageFilename(t *testing.T) {
	target := node.GetTarget([]string{})
	if fn := target.GetImageFilename(); fn != fmt.Sprintf(ImageFilenameTmpl, "arm64") {
		t.Error("wrong image filename")
	}
}

func TestK3sTarget_GetImageFilePath(t *testing.T) {
	sep := string(os.PathSeparator)
	target := node.GetTarget([]string{})
	if fn := target.GetImageFilePath("/tmp/foo"); fn != "/tmp/foo"+sep+fmt.Sprintf(ImageFilenameTmpl, "arm64") {
		t.Error("wrong image file path")
	}
}

func TestNode_Marshal(t *testing.T) {
	sshKey := "~/.ssh/id_rsa"
	password := "secret"
	hostname := "black-pearl"
	ipAddress := "127.0.0.1"
	authType := "ssh-key"
	user := "john"
	arch := "aarch64"

	node := &Node{
		Hostname: hostname,
		Address:  ipAddress,
		Auth: Auth{
			Type:     authType,
			User:     user,
			Password: password,
			SSHKey:   sshKey,
		},
		Arch: arch,
	}

	out1, _ := yaml.Marshal(node)
	str1 := string(out1)

	node2 := &Node{}
	_ = yaml.Unmarshal(out1, node2)

	if actual := node2.Hostname; hostname != actual {
		t.Errorf(msg, hostname, actual)
	}

	if actual := node2.Address; ipAddress != actual {
		t.Errorf(msg, ipAddress, actual)
	}

	if actual := node2.Auth.Type; authType != actual {
		t.Errorf(msg, authType, actual)
	}

	if actual := node2.Auth.SSHKey; sshKey != actual {
		t.Errorf(msg, sshKey, actual)
	}

	if actual := node2.Auth.User; user != actual {
		t.Errorf(msg, user, actual)
	}

	if actual := node2.Auth.Password; password != actual {
		t.Errorf(msg, password, actual)
	}

	out2, _ := yaml.Marshal(node2)
	str2 := string(out2)
	fmt.Println(str2)
	if str1 != str2 {
		t.Errorf("%s\n%s", str1, str2)
	}
}

func TestNodes_GetTargets(t *testing.T) {
	nodes := Nodes{{}, {}, {}}
	sshAuthorizedKeys := []string{}
	targets := nodes.Targets(sshAuthorizedKeys)
	expected := len(nodes)
	actual := len(targets)
	if actual != expected {
		t.Errorf("expected: %d, actual: %d", expected, actual)
	}
}

func TestTargets_SetServerIP(t *testing.T) {
	nodes := Nodes{{}, {}, {}}
	targets := nodes.Targets([]string{})
	expected := "10.0.0.1"
	targets.SetServerIP(expected)

	for _, target := range targets {
		actual := target.ServerIP
		if actual != expected {
			t.Errorf("expected: %s, actual: %s", expected, actual)
		}
	}
}
