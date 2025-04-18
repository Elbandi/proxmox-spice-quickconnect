package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Telmate/proxmox-api-go/proxmox"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/ini.v1"
)

const (
	viewerKey  = "viewer"
	hostKey    = "host"
	userKey    = "user"
	passKey    = "pass"
	vmidKey    = "vmid"
	timeoutKey = "timeout"
)

func CheckErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func CreateContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	signalChan := make(chan os.Signal, 1)

	go func() {
		signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
		defer signal.Stop(signalChan)
		select {
		case <-signalChan:
		case <-ctx.Done():
		}
		cancel()
		signal.Reset()
	}()
	return ctx, cancel
}

func main() {
	configPtr := flag.String("config", "", "config file")
	viewerPathPtr := flag.String(viewerKey, "remote-viewer", "Path to remote-viewer")
	hostPtr := flag.String(hostKey, "", "proxmox host")
	userPtr := flag.String(userKey, "", "proxmox username")
	passPtr := flag.String(passKey, "", "proxmox password")
	vmidPtr := flag.Int(vmidKey, 0, "custom vmid")
	timeoutPtr := flag.Duration(timeoutKey, 10, "custom timeout")
	flag.Parse()
	log.SetOutput(os.Stderr)

	if len(*configPtr) > 0 {
		configPath := *configPtr
		fi, err := os.Stat(configPath)
		CheckErr(err)
		if fi.Size() == 0 {
			log.Fatal("empty file")
		}
		cfg, err := ini.Load(configPath)
		CheckErr(err)
		section, err := cfg.GetSection(ini.DEFAULT_SECTION)
		CheckErr(err)

		SetKeyValue := func(name string, dest *string) error {
			if section.HasKey(name) {
				key, err := section.GetKey(name)
				if err != nil {
					return err
				}
				*dest = key.Value()
			}
			return nil
		}

		CheckErr(SetKeyValue(viewerKey, viewerPathPtr))
		CheckErr(SetKeyValue(hostKey, hostPtr))
		CheckErr(SetKeyValue(userKey, userPtr))
		CheckErr(SetKeyValue(passKey, passPtr))
		var vmid string
		CheckErr(SetKeyValue(vmidKey, &vmid))
		*vmidPtr, err = strconv.Atoi(vmid)
	}
	if len(*hostPtr) == 0 || len(*userPtr) == 0 || *vmidPtr < 1 {
		flag.Usage()
		os.Exit(1)
	}
	if len(*passPtr) == 0 {
		fmt.Print("Enter Password: ")
		bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
		CheckErr(err)
		*passPtr = strings.TrimSpace(string(bytePassword))
	}

	ctx, cancel := CreateContext()
	defer cancel()

	httpClient := http.Client{
		Timeout: *timeoutPtr * time.Second,
	}

	client, err := proxmox.NewClient(fmt.Sprintf("https://%s/api2/json", *hostPtr), &httpClient, "", &tls.Config{InsecureSkipVerify: true}, "", 300)
	CheckErr(err)
	CheckErr(client.Login(ctx, *userPtr, *passPtr, ""))
	vmr := proxmox.NewVmRef(proxmox.GuestID(*vmidPtr))
	config, err := client.GetVmSpiceProxy(ctx, vmr)
	CheckErr(err)

	subProcess := exec.Command(*viewerPathPtr, "-")
	stdin, err := subProcess.StdinPipe()
	CheckErr(err)
	defer stdin.Close()
	//subProcess.Stderr = os.Stderr
	//subProcess.Stdout = os.Stdout

	CheckErr(subProcess.Start())
	_, err = fmt.Fprintf(stdin, "[virt-viewer]\n"+
		"tls-port=%.0f\n"+
		"delete-this-file=%.0f\n"+
		"title=%s\n"+
		"proxy=%s\n"+
		"toggle-fullscreen=%s\n"+
		"type=%s\n"+
		"release-cursor=%s\n"+
		"host-subject=%s\n"+
		"password=%s\n"+
		"secure-attention=%s\n"+
		"host=%s\n"+
		"ca=%s\n",
		config["tls-port"], config["delete-this-file"], config["title"], config["proxy"], config["toggle-fullscreen"],
		config["type"], config["release-cursor"], config["host-subject"], config["password"], config["secure-attention"],
		config["host"], config["ca"])
	CheckErr(err)
	go func() {
		err = subProcess.Wait()
		fmt.Printf("Command finished with error: %v", err)
	}()
}
