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
	target := node.GetK3sTarget([]string{})
	if fn := target.GetImageFilename(); fn != fmt.Sprintf(imageFilenameTmpl, "arm64") {
		t.Error("wrong image filename")
	}
}

func TestK3sTarget_GetImageFilePath(t *testing.T) {
	sep := string(os.PathSeparator)
	target := node.GetK3sTarget([]string{})
	if fn := target.GetImageFilePath("/tmp/foo"); fn != "/tmp/foo"+sep+fmt.Sprintf(imageFilenameTmpl, "arm64") {
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
        Auth:     Auth{
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
