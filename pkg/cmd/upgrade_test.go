package cmd

import (
	"github.com/TheNatureOfSoftware/k3pi/pkg/client"
	"github.com/TheNatureOfSoftware/k3pi/test"
	"testing"
)

func TestUpgradeK3s(t *testing.T) {

	//nodes := model.Nodes{&model.Node{
	//	Hostname: "k3s-node1",
	//	Address:  "192.168.1.128",
	//	Auth: model.Auth{
	//		Type:   model.AuthTypeSSHKey,
	//		User:   "rancher",
	//		SSHKey: "~/.ssh/id_rsa",
	//	},
	//	Arch: "armv7l",
	//}}

	clientFactory.Create = client.NewFakeClient
	nodes := test.CreateNodes()[0:1]

	err := UpgradeK3s("v0.9.1", nodes, false)
	if err != nil {
		t.Error(err)
	}
}
