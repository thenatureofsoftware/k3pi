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

// Package cmd handles k3pi use cases
package cmd

import (
	"github.com/TheNatureOfSoftware/k3pi/pkg/misc"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
	"strings"
)

// SupportedArch supported architectures
var SupportedArch = map[string]bool{
	"armv6l":  true,
	"armv7l":  true,
	"aarch64": true,
}

// ScanRequest parameter type for scanning for nodes
type ScanRequest struct {
	Cidr, HostnameSubString string
	Port int
	SSHAuth *model.Auth
	UserCredentials         map[string]string
}

func (request *ScanRequest) Auths() model.Auths {
	var auths = model.Auths{}
	auths = append(auths, request.SSHAuth)
	for username, password := range request.UserCredentials {
		auths = append(auths, &model.Auth{
			Type:     model.AuthTypeBasicAuth,
			User:     username,
			Password: password,
		})
	}
	return auths
}

func checkArch(address *model.Address, auth *model.Auth) (bool, string) {
	client, err := clientFactory.Create(auth, address)
	if err != nil {
		return false, ""
	}

	result, err := client.Cmd("uname -m").Output()
	if err != nil {
		return false, ""
	}

	arch := strings.TrimSpace(string(result))
	if _, supported := SupportedArch[arch]; supported {
		return supported, arch
	}

	return false, ""
}

func checkIfHostnameMatch(hostnameSubStr string, address *model.Address, auth *model.Auth) (string, bool) {

	client, err := clientFactory.Create(auth, address)
	if err != nil {
		return "", false
	}

	result, err := client.Cmd("hostname").Output()
	if err != nil {
		return "", false
	}

	hostname := strings.TrimSpace(string(result))
	return hostname, strings.Contains(hostname, hostnameSubStr)
}

func ScanForNodes(scanRequest *ScanRequest, hostScanner misc.HostScanner) (*[]model.Node, error) {

	alive, err := hostScanner.ScanForAliveHosts(scanRequest.Cidr)
	if err != nil {
		return nil, err
	}

	var raspberries []model.Node

	for i := range *alive {
		address := model.NewAddress((*alive)[i], scanRequest.Port)
		for _, auth := range scanRequest.Auths() {
			if b, arch := checkArch(&address, auth); b {
				if hn, ok := checkIfHostnameMatch(scanRequest.HostnameSubString, &address, auth); ok {
					raspberries = append(raspberries, model.Node{
						Hostname: hn,
						Address:  address,
						Arch:     arch,
						Auth: *auth,
					})
					break
				}
			}
		}
	}

	return &raspberries, nil
}

