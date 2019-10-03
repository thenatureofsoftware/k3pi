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
package ssh

import (
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg"
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

func TestRunCommand(t *testing.T) {
	t.Skip("manual test")
	settings := &Settings{User: "tnos", KeyPath: "~/.ssh/id_rsa", Port: "22"}
	config, closeHandler := NewClientConfig(settings)

	defer closeHandler()

	ctx := &pkg.CmdOperatorCtx{
		Address:         fmt.Sprintf("%s:%s", "192.168.1.31", settings.Port),
		SSHClientConfig: config,
		EnableStdOut:    false,
	}
	cmdOperator, _ := NewCmdOperator(ctx)
	defer cmdOperator.Close()

	result, err := cmdOperator.Execute("echo hello")
	if err != nil {
		t.Errorf("command execution failed: %d", err)
	}

	if string(result.StdOut) != "hello\n" {
		t.Fail()
	}
}
