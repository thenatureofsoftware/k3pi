package model

import (
	"fmt"
	"github.com/kubernetes-sigs/yaml"
	"testing"
)

var (
	msg  = "\nexpected: %v\nactual: %v"
	node = &Node{
		Arch: "aarch64",
	}
)

func TestNode_Marshal(t *testing.T) {
	sshKey := "~/.ssh/id_rsa"
	password := "secret"
	hostname := "black-pearl"
	ip := "127.0.0.1"
	port := 22
	authType := "ssh-key"
	user := "john"
	arch := "aarch64"

	address := Address{
		IP:   ip,
		Port: port,
	}

	node := &Node{
		Hostname: hostname,
		Address:  address,
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

	if actual := node2.Address; address != actual {
		t.Errorf(msg, address, actual)
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

func TestNodes_IPAddresses(t *testing.T) {

}

func TestNodes_GetTargets(t *testing.T) {
	nodes := Nodes{{}, {}, {}}
	sshAuthorizedKeys := []string{}
	targets := NewK3OSNodes(nodes, sshAuthorizedKeys, "")
	expected := len(nodes)
	actual := len(targets)
	if actual != expected {
		t.Errorf("expected: %d, actual: %d", expected, actual)
	}
}

func TestTargets_SetServerIP(t *testing.T) {
	nodes := Nodes{{}, {}, {}}
	targets := NewK3OSNodes(nodes, []string{}, "")
	expected := "10.0.0.1"
	targets.SetServerIP(expected)

	for _, target := range targets {
		actual := target.ServerIP
		if actual != expected {
			t.Errorf("expected: %s, actual: %s", expected, actual)
		}
	}
}
