package config

import (
	"box/api/digitalocean"
	"box/api/digitalocean/enum/droplet"
	"box/api/digitalocean/enum/region"
	"box/api/digitalocean/sshkeys"
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

const configDirName = ".prototype.do"

// Config holds the project configuration
type Config struct {
	DigitalOceanAPIKey string
	Region             string
	VolumeSize         int
	DropletSlug        string
	BareDomainName     string
	PrivateKeyFilename string
	ImageID            int
	PublicKeyID        int
	BlockStorageID     string
	DropletID          int
	DropletPublicIP    string
}

const defaultRegion = region.NYC3
const defaultDroplet = droplet.S2VCPU2GB

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
	return fmt.Sprintf("%v.yml", projectName)
}

// getTextInput prompts the user for a single line of input, then validates said input.
// if validation fails, then the user is prompted again (this loop continues until input is valid).
func getTextInput(prompt string, validationFunc func(string) bool) (string, error) {
	result := false
	var input string
	var err error
	for result == false {
		fmt.Println(prompt)
		fmt.Print(">")
		reader := bufio.NewReader(os.Stdin)
		input, err = reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		// Trim any leading or trailing whitespace, including the delimiter
		input = strings.Trim(input, " \n\t")
		result = validationFunc(input)
	}

	return input, nil
}

// NewConfig returns a new config struct based on user input
func NewConfig(projectName string) (*Config, error) {
	config := Config{}

	apiKey, err := getTextInput(
		"DigitalOcean API key",
		func(value string) bool {
			return len(value) > 0
		},
	)
	if err != nil {
		return nil, err
	}
	config.DigitalOceanAPIKey = apiKey

	for _, r := range region.Values {
		fmt.Printf("%v	%v\n", r, region.GetName(r))
	}
	r, err := getTextInput(
		fmt.Sprintf("Deployment region (enter for default %v)", defaultRegion),
		func(value string) bool {
			return (len(value) == 0 ||
				region.IsValid(value))
		},
	)
	if err != nil {
		return nil, err
	}

	if len(r) == 0 {
		config.Region = defaultRegion
	} else {
		config.Region = r
	}

	volumeStr, err := getTextInput(
		"Volume size in gigabytes (minimum 1)",
		func(value string) bool {
			i, err := strconv.Atoi(value)
			if err != nil {
				return false
			}

			return i >= 1
		},
	)
	if err != nil {
		return nil, err
	}
	config.VolumeSize, _ = strconv.Atoi(volumeStr)

	for _, d := range droplet.Values {
		fmt.Println(d)
	}
	d, err := getTextInput(
		fmt.Sprintf("Droplet slug (enter for default %v)", defaultDroplet),
		func(value string) bool {
			return (len(value) == 0 ||
				droplet.IsValid(value))
		},
	)
	if err != nil {
		return nil, err
	}
	if len(d) == 0 {
		d = defaultDroplet
	}
	config.DropletSlug = d

	bareDomain, err := getTextInput(
		"Bare domain name (eg: mysite.com)",
		func(value string) bool {
			// Simple regular expression to reject trivial cases, not meant to be thorough
			matched, err := regexp.Match("^\\w+\\.\\w{1,3}$", []byte(value))
			if err != nil {
				panic(err)
			}
			return matched
		},
	)
	if err != nil {
		return nil, err
	}
	config.BareDomainName = bareDomain

	homeDir, _ := os.UserHomeDir()
	defaultKeyFile := path.Join(homeDir, ".ssh", "id_rsa")
	pkFilename, err := getTextInput(
		fmt.Sprintf("Private key full path (Enter for default %v)", defaultKeyFile),
		func(value string) bool {
			filename := value
			if len(value) == 0 {
				filename = defaultKeyFile
			}
			_, err := os.Stat(filename)
			if err != nil {
				fmt.Println("Unable to locate or access keyfile")
				return false
			}

			return true
		},
	)
	if err != nil {
		return nil, err
	}
	if pkFilename == "" {
		pkFilename = defaultKeyFile
	}
	config.PrivateKeyFilename = pkFilename

	// Load public key
	pbkFilename := fmt.Sprintf("%v.pub", pkFilename)
	pbkData, err := ioutil.ReadFile(pbkFilename)
	if err != nil {
		return nil, fmt.Errorf("Unable to read public key file %v", pbkFilename)
	}

	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}
	configFilePath := path.Join(configDir, getConfigFilename(projectName))

	doSvc := digitalocean.NewService(apiKey)
	createdKey, err := sshkeys.Create(doSvc, fmt.Sprintf("box-key-%v", strings.ToLower(projectName)), string(pbkData))
	if err != nil {
		fmt.Println("Unable to post new SSH public key to DigitalOcean")
		return nil, err
	}

	config.PublicKeyID = createdKey.ID

	configBytes, err := yaml.Marshal(&config)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(configFilePath, configBytes, os.FileMode(0600))
	if err != nil {
		return nil, fmt.Errorf("Unable to write configuration file %v", configFilePath)
	}

	return &config, nil
}

// LoadConfig loads the YAML configuration associated with the project name, or returns an error
func LoadConfig(projectName string) (*Config, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	configFilePath := path.Join(configDir, getConfigFilename(projectName))

	configData, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("Unable to read file at %v, are you sure the project exists?", configFilePath)
	}

	config := Config{}
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
