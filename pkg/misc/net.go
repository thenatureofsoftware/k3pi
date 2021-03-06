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

// Package misc miscellaneous functionality
package misc

import (
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
	"github.com/pkg/errors"
	"net"
	"os/exec"
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
	}

	return ips, nil
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
	IP    string
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
		pongChan <- pong{IP: ip, Alive: alive}
	}
}

func receivePong(pongNum int, pongChan <-chan pong, doneChan chan<- []pong) {
	var alive []pong
	for i := 0; i < pongNum; i++ {
		pong := <-pongChan
		//  fmt.Println("received:", pong)
		if pong.Alive {
			alive = append(alive, pong)
		}
	}
	doneChan <- alive
}

// HostScanner scans for hosts
type HostScanner interface {
	ScanForAliveHosts(cidr string) (*[]string, error)
}

// NewHostScanner factory method for a host scanner
func NewHostScanner() HostScanner {
	return &hostScanner{}
}

type hostScanner struct{}

// ScanForAliveHosts scans for all hosts that are alive
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

	var aliveHosts []string
	for _, h := range <-doneChan {
		aliveHosts = append(aliveHosts, h.IP)
	}

	return &aliveHosts, nil
}

// CopyKubeconfig copies kubeconfig from server node
func CopyKubeconfig(kubeconfigFile string, node *model.Node) error {

	out, err := exec.Command(
		"scp",
		"-o",
		"StrictHostKeyChecking=no",
		"-P",
		fmt.Sprintf("%d", node.Address.Port),
		fmt.Sprintf("rancher@%s:/etc/rancher/k3s/k3s.yaml", node.Address.IP),
		kubeconfigFile).CombinedOutput()

	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to copy kubeconfig, %s", out))
	}

	return nil
}
