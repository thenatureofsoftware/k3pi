package model

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	// DefaultK3OSVersion default k3OS version
	DefaultK3OSVersion = "v0.3.0"
	// AuthTypeSSHKey ssh private key authentication
	AuthTypeSSHKey = "ssh-key"
	// AuthTypeBasicAuth username / password authentication
	AuthTypeBasicAuth = "basic-auth"
)

// SSHKeys set of SSH keys
type SSHKeys []string

// RemoteAsset represents an asset that can be downloaded
type RemoteAsset struct {
	Filename, FileURL, CheckSumFilename, CheckSumURL string
}

// RemoteAssets a slice of asssets
type RemoteAssets []*RemoteAsset

// RemoteAssetOwner owner of remote assets
type RemoteAssetOwner interface {
	GetRemoteAssets() RemoteAssets
}

// Task a task to be run for a node set
type Task struct {
	DryRun bool
}

// InstallationTarget target for an installation
type InstallationTarget struct {
}

// Installer an installer that installs
type Installer interface {
	Install() error
}

// Installers a set of installers
type Installers []Installer

// InstallerFactory factory for making installers
type InstallerFactory interface {
	Supports(task interface{}) bool
	MakeInstallers(task interface{}, resourceDir string) Installers
}

// InstallerFactories a set of installer factories
type InstallerFactories interface {
	// GetFactory fetches the installer factory for a install task
	GetFactory(task interface{}) InstallerFactory
}

// InstallerFactoriesT a set of installer factories
type InstallerFactoriesT []InstallerFactory

// GetFactory looks up an installer factory for a given task
func (infs *InstallerFactoriesT) GetFactory(task interface{}) InstallerFactory {
	for _, f := range *infs {
		if f.Supports(task) {
			return f
		}
	}
	return nil
}

// Address address for SSH access
type Address struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

// String address as string <ip>:<port>
func (a Address) String() string {
	return fmt.Sprintf("%s:%d", a.IP, a.Port)
}

// NewAddress creates a new address from ip and port
func NewAddress(ip string, port int) Address {
	return Address{
		IP:   ip,
		Port: port,
	}
}

// NewAddressStr creates a new address from ip and port strings
func NewAddressStr(ip, port string) Address {
	return ParseAddress(fmt.Sprintf("%s:%s", ip, port))
}

// ParseAddress parses an address from a string "<ip>:<port>"
func ParseAddress(s string) Address {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return Address{}
	}

	port, _ := strconv.Atoi(parts[1])
	return Address{
		IP:   parts[0],
		Port: port,
	}
}

// Auth node authentication
type Auth struct {
	Type     string `json:"type"`
	User     string `json:"user"`
	Password string `json:"password,omitempty"`
	SSHKey   string `json:"ssh_key,omitempty"`
}

// Auths authentications
type Auths []*Auth

// Node represents a machine witn an IP and authentication for SSH access
type Node struct {
	Hostname string  `json:"hostname"`
	Address  Address `json:"address"`
	Auth     Auth    `json:"auth"`
	Arch     string  `json:"arch"`
}

// GetArch returns the architecture for the given node. Alternative architecture identifiers can be supplied
// as string separated by : example: arm:armhf
func (n *Node) GetArch(alternatives ...string) string {
	altMap := make(map[string]string)
	for i := range alternatives {
		split := strings.Split(alternatives[i], ":")
		altMap[split[0]] = split[1]
	}
	var arch string
	switch n.Arch {
	case "x86_64":
		arch = "amd64"
	case "armv6l", "armv7l":
		arch = "arm"
	case "aarch64":
		arch = "arm64"
	default:
		return "unknown"
	}
	if alt, ok := altMap[arch]; ok {
		return alt
	}
	return arch
}

// Nodes slice of nodes
type Nodes []*Node

// Info collects info from a set of nodes
func (nodes *Nodes) Info(collect func(node *Node) string) []string {
	var info []string
	for _, v := range *nodes {
		info = append(info, collect(v))
	}
	return info
}

// IPAddresses IP-addresses for a set of nodes
func (nodes *Nodes) IPAddresses() []string {
	var ipAddresses []string
	for _, v := range *nodes {
		ipAddresses = append(ipAddresses, v.Address.IP)
	}
	return ipAddresses
}

// K3OSNode target node for k3OS install
type K3OSNode struct {
	Node
	ServerIP, Token   string
	SSHAuthorizedKeys []string
}

// K3OSNodes k3OS nodes
type K3OSNodes []*K3OSNode

// SetServerIP sets the server ip on all nodes
func (targets *K3OSNodes) SetServerIP(serverIP string) {
	for _, target := range *targets {
		target.ServerIP = serverIP
	}
}

// NewK3OSNode factory method for creating a new k3OS node
func NewK3OSNode(node *Node, sshAuthorizedKeys SSHKeys, token string) *K3OSNode {
	return &K3OSNode{
		Token:             token,
		SSHAuthorizedKeys: sshAuthorizedKeys,
		Node:              *node,
	}
}

// NewK3OSNodes factory method for creating multiple k3OS nodes
func NewK3OSNodes(nodes Nodes, sshAuthorizedKeys []string, token string) K3OSNodes {
	var targets K3OSNodes
	for _, node := range nodes {
		targets = append(targets, NewK3OSNode(node, sshAuthorizedKeys, token))
	}
	return targets
}
