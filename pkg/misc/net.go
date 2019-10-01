package misc

import (
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

type hostScanner struct {}

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
