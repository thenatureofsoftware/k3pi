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
package cmd

import (
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
	"github.com/pkg/errors"
	"testing"
)

func TestSelectServerAndAgents_No_Match(t *testing.T) {
	nodes := []*model.Node{{}, {}, {}, {}}
	server, agents, err := SelectServerAndAgents(nodes, "missing")

	if err != nil {
		t.Error(errors.Wrap(err, "unexpected error"))
	}

	if server != nil {
		t.Errorf("no server expected")
	}

	expectedAgentCount := 4
	if actual := len(agents); actual != expectedAgentCount {
		t.Errorf("expected %d agents, actual: %d", expectedAgentCount, actual)
	}
}

func TestSelectServerAndAgents_No_Nodes(t *testing.T) {
	var nodes []*model.Node
	server, agents, err := SelectServerAndAgents(nodes, "my-server")

	if err != nil {
		t.Error(errors.Wrap(err, "unexpected error"))
	}

	if server != nil {
		t.Errorf("no server expected")
	}

	expectedAgentCount := 0
	if actual := len(agents); actual != expectedAgentCount {
		t.Errorf("expected %d agents, actual: %d", expectedAgentCount, actual)
	}
}

func TestSelectServerAndAgents_Match_Hostname(t *testing.T) {
	hostname := "my-server"
	nodes := []*model.Node{{}, {}, {Hostname: hostname}, {}}
	server, agents, err := SelectServerAndAgents(nodes, hostname)

	if err != nil {
		t.Error(errors.Wrap(err, "unexpected error"))
	}

	if server == nil {
		t.Errorf("expected server is nil")
	}

	expectedAgentCount := 3
	if actual := len(agents); actual != expectedAgentCount {
		t.Errorf("expected %d agents, actual: %d", expectedAgentCount, actual)
	}
}

func TestSelectServerAndAgents_Match_Address(t *testing.T) {
	address := "my-server"
	nodes := []*model.Node{{Address: address}}
	server, agents, err := SelectServerAndAgents(nodes, address)

	if err != nil {
		t.Error(errors.Wrap(err, "unexpected error"))
	}

	if server == nil {
		t.Errorf("expected server is nil")
	}

	expectedAgentCount := 0
	if actual := len(agents); actual != expectedAgentCount {
		t.Errorf("expected %d agents, actual: %d", expectedAgentCount, actual)
	}
}
