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
package misc

import (
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
	"os"
	"testing"
)

func TestHostScanner_ScanForAliveHosts_Localhost(t *testing.T) {
	scanner := NewHostScanner()

	alive, err := scanner.ScanForAliveHosts("127.0.0.1/32")
	if err != nil {
		t.Error(err)
	}

	verifyNumOfHosts(1, len(*alive), t)
}

func TestHostScanner_ScanForAliveHosts_Invalid_Cidr(t *testing.T) {
	scanner := NewHostScanner()

	alive, err := scanner.ScanForAliveHosts("I'm not a CIDR expr")
	if err != nil {
		t.Error(err)
	}

	verifyNumOfHosts(0, len(*alive), t)
}

func verifyNumOfHosts(want int, found int, t *testing.T) {
	if want != found {
		t.Errorf("wanted: %d but found: %d alive hosts", want, found)
	}
}

func TestCopyKubeconfig(t *testing.T) {
	t.Skip("manual test")

	node := &model.Node{
		Address: model.ParseAddress("192.168.1.128:22"),
	}

	fn := CreateTempFileName(os.TempDir(), "k3s-*.yaml")
	defer os.RemoveAll(fn)

	err := CopyKubeconfig(fn, node)

	if err != nil {
		t.Error(err)
	}
}
