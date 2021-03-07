package main

import (
	"log"

	"github.com/alecthomas/kong"
)

var cli struct {
}

func main() {
	ctx := kong.Parse(&cli)
	err := ctx.Run()
	if err != nil {
		log.Fatalln(err)
	}
}
