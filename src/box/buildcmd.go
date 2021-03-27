package main

import (
	"box/config"
	"box/manifest"
	"box/runtime"
	"fmt"
	"os"
	"path/filepath"
)

type BuildCmd struct {
}

func (cmd *BuildCmd) Run() error {
	dirName, err := os.Getwd()
	if err != nil {
		return err
	}

	manifestFilename := filepath.Join(dirName, "box.yml")
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

	err = rt.Build()

	if err != nil {
		return err
	}

	fmt.Println("Build complete!")
	return nil
}
