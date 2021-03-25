package main

import (
	"box/manifest"
	"box/runtime"
	"os"
	"path/filepath"
)

type ShutdownCmd struct {
}

func (cmd *ShutdownCmd) Run() error {
	dirName, err := os.Getwd()
	if err != nil {
		return err
	}

	manifestFilename := filepath.Join(dirName, "box.yml")
	mfst, err := manifest.NewManifest(manifestFilename)
	if err != nil {
		return err
	}

	// Production is irrelevant for the shutdown command, behavior is always the same
	rt, err := runtime.New(mfst, nil, false)
	if err != nil {
		return err
	}

	rt.Shutdown()

	return nil
}
