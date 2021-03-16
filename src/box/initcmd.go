package main

import (
	"box/config"
)

type InitCmd struct {
	Name string `arg help="Project name"`
}

func (cmd *InitCmd) Run() error {
	_, err := config.New(cmd.Name)
	if err != nil {
		return err
	}

	return nil
}
