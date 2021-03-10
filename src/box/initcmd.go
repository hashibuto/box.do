package main

import (
	"fmt"
	"os"

	"box/config"
)

type InitCmd struct {
	Name string `arg help="Project name"`
}

func (cmd *InitCmd) Run() error {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return err
	}

	// Read, write, execute by this user only
	err = os.MkdirAll(configDir, os.FileMode(0700))
	if err != nil {
		return err
	}

	config, err := config.NewConfig(cmd.Name)
	if err != nil {
		return err
	}

	fmt.Println(config)
	return nil
}
