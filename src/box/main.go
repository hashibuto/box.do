package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

var cli struct {
	Init     InitCmd       `cmd help="Initializes a new project"`
	Mkimage  MkImageCmd    `cmd help="Make DigitalOcean base image"`
	Mkremote MakeRemoteCmd `cmd help="Provision remote host and respective resources"`
}

func main() {
	ctx := kong.Parse(&cli)
	err := ctx.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
