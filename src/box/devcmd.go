package main

import (
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
	mfst, err := manifest.NewManifest(manifestFilename)
	if err != nil {
		return err
	}

	rt, err := runtime.New(mfst, false)
	if err != nil {
		return err
	}

	fmt.Println(rt)

	return nil
}
