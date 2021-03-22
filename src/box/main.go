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
	Dev      DevCmd        `cmd help="Run box project in development mode"`
	Shutdown ShutdownCmd   `cmd help="Shut down the current project"`
}

func main() {
	ctx := kong.Parse(&cli)
	err := ctx.Run()
	if err != nil {
		fmt.Printf("\n%v\n", err)
		os.Exit(1)
	}
}
