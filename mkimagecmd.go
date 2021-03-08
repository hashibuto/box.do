package main

import (
	"fmt"
	"os"
	"prdo/config"
)

type MkImageCmd struct {
	Name      string `arg help="Project name"`
	Overwrite bool   `default=false help="Overwrite existing image"`
}

func (cmd *MkImageCmd) Run() error {
	cfg, err := config.LoadConfig(cmd.Name)
	if err != nil {
		return err
	}

	if cmd.Overwrite == false && len(cfg.ImageID) != 0 {
		fmt.Println("It appears you've already created a DigitalOcean image.")
		fmt.Println("Please use the --overwrite option if you'd like to overwrite it with a new one.")
		os.Exit(1)
	}

	return nil
}
