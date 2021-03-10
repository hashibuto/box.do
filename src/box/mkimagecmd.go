package main

import (
	"box/api/digitalocean"
	"box/api/digitalocean/droplet"
	dropletenum "box/api/digitalocean/enum/droplet"
	"box/config"
	"box/sshconn"
	"fmt"
	"os"
	"time"
)

const imageHostName = "box-image-maker"
const maxConnectAttempts = 10
const sshRetrySeconds = 10

type MkImageCmd struct {
	Name      string `arg help="Project name"`
	Overwrite bool   `default=false help="Overwrite existing image"`
}

func (cmd *MkImageCmd) Run() error {
	var dropletObj *droplet.Droplet
	cfg, err := config.LoadConfig(cmd.Name)
	if err != nil {
		return err
	}

	if cmd.Overwrite == false && cfg.ImageID != 0 {
		fmt.Println("It appears you've already created a DigitalOcean image.")
		fmt.Println("Please use the --overwrite option if you'd like to overwrite it with a new one.")
		os.Exit(1)
	}

	fmt.Println("Creating new droplet for base image...")
	doSvc := digitalocean.NewService(cfg.DigitalOceanAPIKey)
	dropletObj, err = droplet.CreateFromPublicImage(
		doSvc,
		imageHostName,
		dropletenum.S1VCPU1GB, // Minimum size for image build
		cfg.Region,
		droplet.DefaultPublicImage,
		[]int{cfg.PublicKeyID},
	)

	for dropletObj.Status == "new" {
		time.Sleep(time.Second * 5)
		fmt.Println("Checking droplet status...")
		dropletObj, err = droplet.GetByID(doSvc, dropletObj.ID)
	}

	if dropletObj.Status != "active" {
		fmt.Printf("Droplet creation failed, please be sure to remove the droplet named %v from your DigitalOcean control panel\n", imageHostName)
		os.Exit(1)
	}

	fmt.Println("Droplet successfully created")

	var ipAddress string
	for _, address := range dropletObj.Networks.V4 {
		if address.Type == "public" {
			ipAddress = address.IPAddress
		}
	}

	if ipAddress == "" {
		fmt.Println("Unable to obtain the public IP4 address from the droplet object, aborting")
		os.Exit(1)
	}

	connected := false
	var conn *sshconn.SSHConn
	for i := 0; i < maxConnectAttempts; i++ {
		fmt.Println("Trying to contact via SSH...")
		conn, err = sshconn.NewSSHConn(cfg.PrivateKeyFilename, "root", ipAddress)
		if err == nil {
			connected = true
			break
		}
		fmt.Printf("Failed, trying again in %vs...\n", sshRetrySeconds)
		time.Sleep(time.Second * sshRetrySeconds)
	}

	if connected == false {
		fmt.Printf("Unable to connect via SSH, aborting")
		os.Exit(1)
	}

	defer conn.Close()
	err = conn.Run([]string{":"})
	if err != nil {
		return err
	}

	fmt.Println("SSH is working, continuing")

	return nil
}
