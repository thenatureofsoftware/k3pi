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

package client

import (
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
	"reflect"
	"strings"
	"testing"
)

var (
	ManualTestAddress = model.NewAddress("192.168.1.111", 22)
)

var auth = &model.Auth{
	Type:   model.AuthTypeSSHKey,
	User:   "rancher",
	SSHKey: "~/.ssh/id_rsa",
}

func TestNewClientManual(t *testing.T) {
	t.Skip("manual test")

	c, err := NewClient(auth, &ManualTestAddress)

	if err != nil {
		t.Error(err)
	}

	if reflect.TypeOf(c) != reflect.TypeOf(&client{}) {
		t.Error("wrong type")
	}
}

func TestClientManual_Cmd(t *testing.T) {
	t.Skip("manual test")

	c, err := NewClient(auth, &ManualTestAddress)
	if err != nil {
		t.Error(err)
	}

	out, err := c.Cmd("whoami").Output()
	if err != nil {
		t.Error(err)
	}
	expected := auth.User
	actual := strings.TrimSpace(string(out))

	if expected != actual {
		t.Error(fmt.Sprintf("expected: %s, actual: %s", expected, actual))
	}
}

func TestClientManual_Copy(t *testing.T) {
	t.Skip("manual test")
	c, err := NewClient(auth, &ManualTestAddress)
	if err != nil {
		t.Error(err)
	}

	err = c.Copy("./README.md", "~/README.md")

	if err != nil {
		t.Error(err)
	}
}
