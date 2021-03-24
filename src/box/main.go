package main

import (
	"fmt"
	"os"
	"os/user"

	"github.com/alecthomas/kong"
)

var cli struct {
	Init     InitCmd       `cmd help="Initializes a new project"`
	Mkimage  MkImageCmd    `cmd help="Make DigitalOcean base image"`
	Mkremote MakeRemoteCmd `cmd help="Provision remote host and respective resources"`
	Dev      DevCmd        `cmd help="Run box project in development mode"`
	Shutdown ShutdownCmd   `cmd help="Shut down the current project"`
	Build    BuildCmd      `cmd help="Build the current project"`
}

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	if user.Uid == "0" {
		fmt.Println("box must not be run as root")
		os.Exit(1)
	}

	ctx := kong.Parse(&cli)
	err = ctx.Run()
	if err != nil {
		fmt.Printf("\n%v\n", err)
		os.Exit(1)
	}
}
