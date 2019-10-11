package model

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	DefaultK3OSVersion = "v0.3.0"
	// AuthTypeSSHKey ssh private key authentication
	AuthTypeSSHKey = "ssh-key"
	// AuthTypeBasicAuth username / password authentication
	AuthTypeBasicAuth = "basic-auth"
)

// Set of SSH keys
type SSHKeys []string

// Represents an asset that can be downloaded
type RemoteAsset struct {
	Filename, FileUrl, CheckSumFilename, CheckSumUrl string
}

// A slice of asssets
type RemoteAssets []*RemoteAsset

// Owner of remote assets
type RemoteAssetOwner interface {
	GetRemoteAssets() RemoteAssets
}

// A task to be run for a node set
type Task struct {
	DryRun bool
}

// Target for an installation
type InstallationTarget struct {
}

// An installer that installs
type Installer interface {
	Install() error
}

// A set of installers
type Installers []Installer

// Factory for making installers
type InstallerFactory interface {
	Supports(task interface{}) bool
	MakeInstallers(task interface{}, resourceDir string) Installers
}

// A set of installer factories
type InstallerFactories []InstallerFactory

// Fetches the installer factory for a install task
func (inf *InstallerFactories) GetFactory(task interface{}) InstallerFactory {
	for _, f := range *inf {
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

// NewAddress creates a new address from ip and port strings
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

// Node authentication
type Auth struct {
	Type     string `json:"type"`
	User     string `json:"user"`
	Password string `json:"password,omitempty"`
	SSHKey   string `json:"ssh_key,omitempty"`
}

// Represents a machine witn an IP and authentication for SSH access
type Node struct {
	Hostname string  `json:"hostname"`
	Address  Address `json:"address"`
	Auth     Auth    `json:"auth"`
	Arch     string  `json:"arch"`
}

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

// Slice of nodes
type Nodes []*Node

func (nodes *Nodes) Info(collect func(node *Node) string) []string {
	var info []string
	for _, v := range *nodes {
		info = append(info, collect(v))
	}
	return info
}

// IP-addresses for a set of nodes
func (nodes *Nodes) IPAddresses() []string {
	var ipAddresses []string
	for _, v := range *nodes {
		ipAddresses = append(ipAddresses, v.Address.IP)
	}
	return ipAddresses
}

// Target node for k3OS install
type K3OSNode struct {
	Node
	ServerIP, Token   string
	SSHAuthorizedKeys []string
}

type K3OSNodes []*K3OSNode

func (targets *K3OSNodes) SetServerIP(serverIP string) {
	for _, target := range *targets {
		target.ServerIP = serverIP
	}
}

func NewK3OSNode(node *Node, sshAuthorizedKeys SSHKeys, token string) *K3OSNode {
	return &K3OSNode{
		Token:             token,
		SSHAuthorizedKeys: sshAuthorizedKeys,
		Node:              *node,
	}
}

func NewK3OSNodes(nodes Nodes, sshAuthorizedKeys []string, token string) K3OSNodes {
	var targets K3OSNodes
	for _, node := range nodes {
		targets = append(targets, NewK3OSNode(node, sshAuthorizedKeys, token))
	}
	return targets
}
