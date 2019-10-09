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
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
	"github.com/TheNatureOfSoftware/k3pi/pkg/ssh"
	"net"
	"os/exec"
	"time"
)

func hosts(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}
	// remove network address and broadcast address
	if len(ips) > 3 {
		return ips[1 : len(ips)-1], nil
	} else {
		return ips, nil
	}
}

//  http://play.golang.org/p/m8TNTtygK0
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

type pong struct {
	Ip    string
	Alive bool
}

func ping(pingChan <-chan string, pongChan chan<- pong) {
	for ip := range pingChan {
		_, err := exec.Command("ping", "-c1", "-t1", ip).Output()
		var alive bool
		if err != nil {
			alive = false
		} else {
			alive = true
		}
		pongChan <- pong{Ip: ip, Alive: alive}
	}
}

func receivePong(pongNum int, pongChan <-chan pong, doneChan chan<- []pong) {
	var alives []pong
	for i := 0; i < pongNum; i++ {
		pong := <-pongChan
		//  fmt.Println("received:", pong)
		if pong.Alive {
			alives = append(alives, pong)
		}
	}
	doneChan <- alives
}

type HostScanner interface {
	ScanForAliveHosts(cidr string) (*[]string, error)
}

func NewHostScanner() HostScanner {
	return &hostScanner{}
}

type hostScanner struct{}

func (h *hostScanner) ScanForAliveHosts(cidr string) (*[]string, error) {
	hosts, _ := hosts(cidr)
	concurrentMax := 50
	pingChan := make(chan string, concurrentMax)
	pongChan := make(chan pong, len(hosts))
	doneChan := make(chan []pong)

	for i := 0; i < concurrentMax; i++ {
		go ping(pingChan, pongChan)
	}

	go receivePong(len(hosts), pongChan, doneChan)

	for _, ip := range hosts {
		pingChan <- ip
	}

	aliveHosts := []string{}
	for _, h := range <-doneChan {
		aliveHosts = append(aliveHosts, h.Ip)
	}

	return &aliveHosts, nil
}

func WaitForNode(node *model.Node, sshSettings *ssh.Settings, timeout time.Duration) error {

	resolvedSSHSettings := resolveSSHSettings(sshSettings)

	clientConfig, sshAgentCloseHandler, err := ssh.NewClientConfig(resolveSSHSettings(sshSettings))
	PanicOnError(err, "failed to create ssh agent")
	defer sshAgentCloseHandler()

	ctx := &pkg.CmdOperatorCtx{
		Address:         fmt.Sprintf("%s:%s", node.Address, resolvedSSHSettings.Port),
		SSHClientConfig: clientConfig,
		EnableStdOut:    false,
	}

	timeToStop := time.Now().Add(timeout)
	for {
		_, err := ssh.NewCmdOperator(ctx)
		if err == nil {
			break
		} else if time.Now().After(timeToStop) {
			return fmt.Errorf("timeout waiting for node: %s", node.Address)
		}
		time.Sleep(time.Second * 2)
	}

	return nil
}

func resolveSSHSettings(sshSettings *ssh.Settings) *ssh.Settings {
	if sshSettings != nil {
		return sshSettings
	}
	return &ssh.Settings{
		User:    "rancher",
		KeyPath: "~/.ssh/id_rsa",
		Port:    "22",
	}
}

func CopyKubeconfig(kubeconfigFile string, node *model.Node) error {
	return exec.Command(
		"scp",
		"-o",
		"StrictHostKeyChecking=no",
		fmt.Sprintf("rancher@%s:/etc/rancher/k3s/k3s.yaml", node.Address),
		kubeconfigFile).Run()
}
