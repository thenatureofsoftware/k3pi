package net

import (
	"fmt"
	ssh2 "github.com/TheNatureOfSoftware/k3pi/pkg/ssh"
	"golang.org/x/crypto/ssh"
	"log"
	"net"
	"os/exec"
	"strings"
)

func Hosts(cidr string) ([]string, error) {
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

type Pong struct {
	Ip    string
	Alive bool
}

func ping(pingChan <-chan string, pongChan chan<- Pong) {
	for ip := range pingChan {
		_, err := exec.Command("ping", "-c1", "-t1", ip).Output()
		var alive bool
		if err != nil {
			alive = false
		} else {
			alive = true
		}
		pongChan <- Pong{Ip: ip, Alive: alive}
	}
}

func receivePong(pongNum int, pongChan <-chan Pong, doneChan chan<- []Pong) {
	var alives []Pong
	for i := 0; i < pongNum; i++ {
		pong := <-pongChan
		//  fmt.Println("received:", pong)
		if pong.Alive {
			alives = append(alives, pong)
		}
	}
	doneChan <- alives
}

func ScanForRaspberries(cidr string, substr string, settings *ssh2.Settings) ([]string, error) {
	hosts, _ := Hosts(cidr)
	concurrentMax := 100
	pingChan := make(chan string, concurrentMax)
	pongChan := make(chan Pong, len(hosts))
	doneChan := make(chan []Pong)

	for i := 0; i < concurrentMax; i++ {
		go ping(pingChan, pongChan)
	}

	go receivePong(len(hosts), pongChan, doneChan)

	for _, ip := range hosts {
		pingChan <- ip
	}

	config, closeHandler, err := ssh2.NewClientConfig(settings)
	if err != nil {
		log.Fatalf("failed to create ssh config: %d", err)
	}
	defer closeHandler()

	raspberries := []string{}
	alive := <-doneChan
	for i := range alive {
		ip := alive[i].Ip
		// fmt.Printf("Checking if %s is a member of the Raspberries: ", ip)
		address := fmt.Sprintf("%s:%s", ip, settings.Port)
		if checkIfRaspberry(address, config) {
			// fmt.Print("Yes\n")
			if checkIfHostnameMatch(address, substr, config) {
				raspberries = append(raspberries, ip)
			}
		} else {
			// fmt.Print("No\n")
		}
	}

	return raspberries, nil
}

func checkIfHostnameMatch(ipAddress string, substr string, config *ssh.ClientConfig) bool {
	if len(substr) == 0 {
		return true
	}
	cmdOperator, err := ssh2.NewCmdOperator(ipAddress, config, false)
	if err != nil {
		return false
	}

	result, err := cmdOperator.Execute("hostname")
	if err != nil {
		return false
	}

	return strings.Contains(string(result.StdOut), substr)
}

func checkIfRaspberry(ipAddress string, config *ssh.ClientConfig) bool {
	cmdOperator, err := ssh2.NewCmdOperator(ipAddress, config, false)
	if err != nil {
		return false
	}

	result, err := cmdOperator.Execute("uname -m")
	if err != nil {
		return false
	}

	return strings.Contains(string(result.StdOut), "armv7l")
}
