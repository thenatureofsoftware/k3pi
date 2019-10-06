# k3pi `/'ki 'pai/`

Tool for installing [k3os](https://github.com/rancher/k3os) on your favorite ARM device.

1. Start your ARM device with an OS image that has `ssh` enabled.
  
   For Raspberry Pi we recommend [HypriotOS](https://blog.hypriot.com/post/releasing-HypriotOS-1-11/). Install instructions can be found [here](https://github.com/hypriot/image-builder-rpi/releases).

2. Scan your network for all your devices that should be part [`k3s`](https://github.com/rancher/k3s) cluster.
   
   ```shell script
   $ k3pi scan --auth pirate:hypriot --substr pearl
   $ # Save the output
   $ $ k3pi scan --auth pirate:hypriot --substr pearl > nodes.yaml
   ```

3. Install `k3os` using the `install` command

   ```shell script
   $ # Make a dry-run install first
   $ k3pi install --filename nodes.yaml --server <your selected server node ip> --dry-run
   $ # You can also combine scan and install
   $ k3pi scan --auth pirate:hypriot --substr pearl | k3pi install -y --server <your selected server node ip> --dry-run
   $ # If every thing looks good, the run the install, this will overwrite your nodes
   $ k3pi install --filename nodes.yaml --server <your selected server node ip>
   ```

#### `scan`
```
$ k3pi scan -h
Scans the network for ARM devices with ssh enabled. The scan can use one SSH key
and multiple username and password combinations. Examples:

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
 Scan and install, confirm the install using --yes
 $ k3pi scan <scan args> | k3pi install --yes <install args>

 You should always run the install as a dry run first
 $ k3pi scan <scan args> | k3pi install <install args> --dry-run

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
