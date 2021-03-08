package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
)

const configDirName = ".prototype.do"

type Config struct {
	DigitalOceanAPIKey string
	Region             string
	VolumeSize         string
	DropletSlug        string
	BareDomainName     string
}

// GetConfigDir returns the configuration directory
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.New("Unable to determine user's home directory")
	}

	fullConfigPath := path.Join(homeDir, configDirName)
	return fullConfigPath, nil
}

func getConfigFilename(projectName string) string {
	return fmt.Sprint("%v.%v", projectName, ".yml")
}

func getTextInput(prompt string, validationFunc func(string) bool) (string, error) {
	result := false
	var input string
	var err error
	for result == false {
		fmt.Print(prompt)
		reader := bufio.NewReader(os.Stdin)
		input, err = reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		result = validationFunc(input)
	}

	return input, nil
}

func NewConfig(projectName string) (*Config, error) {
	config := Config{}

	apiKey, err := getTextInput(
		"Please enter your DigitalOcean API key\n>",
		func(value string) bool {
			return len(value) > 0
		},
	)

	if err != nil {
		return nil, err
	}

	config.DigitalOceanAPIKey = apiKey

	// configDir, err := GetConfigDir()
	// if err != nil {
	// 	return nil, err
	// }
	// configFilePath := path.Join(configDir, getConfigFilename(projectName))

	return &config, nil
}
