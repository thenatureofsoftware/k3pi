[![pipeline status](https://gitlab.com/TheNatureOfSoftware/k3pi/badges/master/pipeline.svg)](https://gitlab.com/TheNatureOfSoftware/k3pi/commits/master) [![Go Report Card](https://goreportcard.com/badge/github.com/TheNatureOfSoftware/k3pi)](https://goreportcard.com/report/github.com/TheNatureOfSoftware/k3pi)

# k3pi `/'ki 'pai/`

Tool for installing [`k3OS`](https://github.com/rancher/k3os) on your favorite ARM device.

## Why

The easiest way to get started with [`k3s`](https://github.com/rancher/k3s) is probably by using
[k3d](https://github.com/rancher/k3s) and Docker. If you wan't to run `k3s` on multiple nodes then the quickest way
is to use Alex Elis [`k3sup`](https://github.com/alexellis/k3sup). But if you wan't to run an OS targeted for `k3s`
like `k3OS` that's where this tool might come in handy.

## Get started

1. Boot your ARM device with an OS image that has `ssh` enabled.
  
   For Raspberry Pi we recommend [Ubuntu](https://ubuntu.com/download/raspberry-pi).

2. Scan your network for all your devices that should be part of the `k3s` cluster.
   
   ```shell script
   $ k3pi scan --auth ubuntu:ubuntu --substr ubuntu
   $ # Save the output
   $ $ k3pi scan --auth ubuntu:ubuntu --substr ubuntu > nodes.yaml
   ```

3. Install `k3os` using the `install` command

   ```shell script
   $ # Make a dry-run install first
   $ k3pi install --filename nodes.yaml --server <your selected server node ip> --dry-run
   $ # You can also combine scan and install
   $ k3pi scan --auth ubuntu:ubuntu --substr pearl | k3pi install -y --server <your selected server node ip> --dry-run
   $ # If every thing looks good, the run the install, this will overwrite your nodes
   $ k3pi install --filename nodes.yaml --server <your selected server node ip>
   ```
## Gotchas

* When installing, if your environment assign a new IP address after reboot you will get an error.
  The installation was probably successful but you need to copy your `kubeconfig` your self.
* If you saved your nodes to file before installing then you need to re-scan your nodes. All nodes
  will have `rancher` as user.
 
## Commands

* [`scan`](#scan) - for finding your target nodes
* [`install`](#install) - for installing k3OS
* [`template`](#template) - for generating sample templates for server and agent

#### `scan`
```
$ k3pi scan -h
Scans the network for ARM devices with ssh enabled. The scan can use one SSH key
and multiple username and password combinations.

        Examples:

        # Scan using default SSH key in ~/.ssh/id_rsa, user root and CIDR 192.168.1.0/24
        $ k3pi scan

        # Scan using default SSH key in ~/.ssh/id_rsa and user foo
        $ k3pi scan --user foo --cidr 192.168.1.0/24

        # Scan using username and password
        $ k3pi scan --auth foo:bar --auth root:notsosecret

        # Scan filtering on hostname
        $ k3pi scan --substr pearl

Usage:
  k3pi scan [flags]

Flags:
  -a, --auth strings     Username and password separated with ':' for authentication
      --cidr string      CIDR to scan for members (default "192.168.1.0/24")
  -h, --help             help for scan
      --ssh-key string   ssh key to use for remote login (default "~/.ssh/id_rsa")
      --ssh-port int     port on which to connect for ssh (default 22)
      --substr string    Substring that should be part of hostname
      --user string      username for ssh login (default "root")
```

#### `install`

```
Installs k3os on ARM devices, should be combined with the scan command.

        IMPORTANT! This will overwrite your existing installation.
        
        Examples:
        
        You should always run the install as a dry run first
        $ k3pi scan <scan args> | k3pi install --yes <install args> --dry-run
        
        Scan and install, confirm the install using --yes
        $ k3pi scan <scan args> | k3pi install --yes <install args>

        Installs k3os on all nodes in the file and selects <server ip> as server
        $ k3pi install --filename ./nodes.yaml --server <server ip>

        $ Installs k3os on all nodes as agents joining an existing server (server is not in nodes file)
        k3pi install --filename ./nodes.yaml -t <token|secret> --server <server ip>

Usage:
  k3pi install [flags]

Flags:
      --dry-run                   if true will run the install but not execute commands
  -f, --filename string           scan output file with all nodes
  -h, --help                      help for install
      --hostname-pattern string   hostname pattern, printf with %s and %d (default "%s%d")
      --hostname-prefix string    hostname prefix, (hostname = '<prefix><index>') (default "k3-node")
  -s, --server string             ip address or hostname of the server node
  -k, --ssh-key strings           ssh authorized key that should be added to the rancher user (default [~/.ssh/id_rsa.pub])
  -t, --token string              token or cluster secret for joining a server
  -y, --yes                       confirm the installation
```

#### `template`

```
Shows the k3OS config template for both server and agent.

        Examples:

        Shows both server and agent config template
        $ k3pi template

Usage:
  k3pi template [flags]

Flags:
  -h, --help   help for template
```

## Links

* [Ubuntu for RaspberryPi](https://ubuntu.com/download/raspberry-pi)
* [Rancher k3OS](https://github.com/rancher/k3os)
* [Rancher k3s](https://github.com/rancher/k3s)
* [Rancher k3d](https://github.com/rancher/k3d)
* [Alex Ellis's k3sup](https://github.com/alexellis/k3sup)
