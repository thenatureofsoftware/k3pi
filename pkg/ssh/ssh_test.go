package ssh

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"testing"
)

func TestCreateSshSettings(t *testing.T) {
	sshSettings := &Settings{KeyPath: "", Port: "22", User: ""}
	if sshSettings == nil {
		t.Fail()
	}
}

func TestLoadPublicKey(t *testing.T) {
	dir, err := ioutil.TempDir(".", "test-")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	keyFile := dir + "/id_rsa"
	app := "ssh-keygen"
	cmd := exec.Command(app, "-b", "2048", "-t", "rsa", "-f", keyFile, "-q", "-N", "")
	stdout, err := cmd.Output()

	if err != nil {
		log.Println(string(stdout))
		t.Error(err.Error())
		return
	}

	sshSettings := &Settings{KeyPath: keyFile}
	publicKey, closeHandler, err := LoadPublicKey(sshSettings)
	if closeHandler == nil {
		t.Error("close handler is nil")
	} else if err != nil {
		t.Errorf("load public key failed: %d", err)
	}
	if publicKey == nil {
		t.Error("public key is nil")
	}
}

func _TestRunCommand(t *testing.T) {
	settings := &Settings{User: "tnos", KeyPath: "~/.ssh/id_rsa", Port: "22"}
	config, closeHandler, err := NewClientConfig(settings)

	defer closeHandler()

	if err != nil {
		t.Errorf("failed to run ping command: %d", err)
	}

	address := fmt.Sprintf("%s:%s", "192.168.1.31", settings.Port)
	cmdOperator, err := NewCmdOperator(address, config)
	defer cmdOperator.Close()

	if err != nil {
		t.Errorf("failed to create cmdOperator: %d", err)
	}

	result, err := cmdOperator.Execute("echo hello")
	if err != nil {
		t.Errorf("command execution failed: %d", err)
	}

	if string(result.StdOut) != "hello\n" {
		t.Fail()
	}
}
