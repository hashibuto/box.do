package main

import (
	"box/config"
	"box/manifest"
	"box/runtime"
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

	manifestFilename := path.Join(dirName, "box.yml")
	fmt.Println("Loading run manifest")
	mfst, err := manifest.NewManifest(manifestFilename)
	if err != nil {
		return err
	}

	// Config file has to exist before a dev project can be started
	fmt.Println("Loading project configuration")
	cfg, err := config.Load(mfst.Project)
	if err != nil {
		return err
	}

	rt, err := runtime.New(mfst, cfg, false)
	if err != nil {
		return err
	}

	err = rt.Start()

	return err
}
