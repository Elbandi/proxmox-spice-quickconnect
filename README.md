# proxmox-spice-quickconnect
Proxmox spice quickconnect

## How to build executable on Linux
Build and install: `go get -u github.com/Elbandi/proxmox-spice-quickconnect`

Download and build only: `wget https://raw.githubusercontent.com/Elbandi/proxmox-spice-quickconnect/master/main.go` and `go build -o proxmox-vm-connect .`

Build command for a windows executable: `env GOOS=windows GOARCH=amd64 go build -o proxmox-vm-connect.exe .`

[Source](https://www.digitalocean.com/community/tutorials/how-to-build-go-executables-for-multiple-platforms-on-ubuntu-16-04)

---

## Command line parameters

`-host pvenodeip -user foobar@pve -pass secret -vmid 123`

or from a config file:

```
host=pvenodeip
user=foobar@pve
pass=secret
vmid=123
viewer=path/to/remote-viewer.exe
```

use: `-config path/to/configfile`

[Source](https://forum.proxmox.com/threads/remote-spice-access-without-using-web-manager.16561/post-255078)


---

