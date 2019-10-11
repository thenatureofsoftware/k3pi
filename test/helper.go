package test

import "github.com/TheNatureOfSoftware/k3pi/pkg/model"

// CreateNodes creates nodes for test
func CreateNodes() model.Nodes {
	n1 := model.Node{
		Hostname: "node1",
		Address:  "10.0.0.1",
		Auth: model.Auth{
			Type:   model.AuthTypeSSHKey,
			User:   "test",
			SSHKey: "~/.ssh/id_rsa",
		},
		Arch: "aarch64",
	}
	n2 := n1
	n2.Hostname = "node2"
	n2.Address = "10.0.0.2"

	n3 := n1
	n3.Hostname = "node3"
	n3.Address = "10.0.0.3"

	return model.Nodes{
		&n1,
		&n2,
		&n3,
	}
}
