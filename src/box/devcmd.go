package main

import (
	"fmt"
	"os"
	"path"
)

type DevCmd struct {
}

func (cmd *DevCmd) Run() error {
	dirName, err := os.Getwd()
	if err != nil {
		return err
	}

	if _, err := os.Stat(path.Join(dirName, "box.yml")); err != nil {
		fmt.Println("Unable to access box.yml at the current location, try running again from your project directory.")
		os.Exit(1)
	}

	return nil
}
