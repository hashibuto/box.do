package config

import (
	"box/api/digitalocean"
	"box/api/digitalocean/enum/droplet"
	"box/api/digitalocean/enum/region"
	"box/api/digitalocean/sshkeys"
	"bufio"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

const configDirName = ".box.do"
const configFileName = "config.yml"
const dataDirName = "data"

var ProjectNameRe *regexp.Regexp = regexp.MustCompile("^[a-z][a-z0-9\\-]{1,18}[a-z0-9]$")

// Config holds the project configuration
type Config struct {
	ProjectName        string
	Email              string
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
	FirewallID         string
	projNameHash       string
}

const defaultRegion = region.NYC3
const defaultDroplet = droplet.S2VCPU2GB

// https://golangcode.com/validate-an-email-address/
var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// CheckProjectExists returns a boolean, indicating whether or not the project exists
func CheckProjectExists(projectName string) (bool, error) {
	dirName, err := GetConfigDir()
	if err != nil {
		return false, fmt.Errorf("config.CheckProjectExists: %w", err)
	}

	projectDir := filepath.Join(dirName, projectName)
	if _, err := os.Stat(projectDir); err != nil {
		return false, nil
	}

	return true, nil
}

func isEmailValid(e string) bool {
	if len(e) < 3 && len(e) > 254 {
		return false
	}
	return emailRegex.MatchString(e)
}

// GetConfigDir returns the configuration directory
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.New("Unable to determine user's home directory")
	}

	fullConfigPath := filepath.Join(homeDir, configDirName)
	return fullConfigPath, nil
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

// New returns a new config struct based on user input
func New(projectName string) (*Config, error) {
	if !ProjectNameRe.Match([]byte(projectName)) {
		return nil, fmt.Errorf(
			"Project name must only contain lowercase alpha, numbers, or hyphens.  It must begin with an alpha character, and may not end with a hyphen.  Project names must be between 3 and 20 characters long.",
		)
	}

	exists, err := CheckProjectExists(projectName)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("A project already exists by the name: %v", projectName)
	}

	config := Config{ProjectName: projectName}

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

	email, err := getTextInput(
		"TLS certificate registration email for use with letsencrypt",
		isEmailValid,
	)
	if err != nil {
		return nil, err
	}
	config.Email = email

	homeDir, _ := os.UserHomeDir()
	defaultKeyFile := filepath.Join(homeDir, ".ssh", "id_rsa")
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

	dataDir := filepath.Join(configDir, projectName, dataDirName)
	os.MkdirAll(dataDir, os.FileMode(0755))

	configFilePath := filepath.Join(
		configDir,
		projectName,
		configFileName,
	)

	doSvc := digitalocean.NewService(apiKey)
	fmt.Printf("Checking for an existing matching SSH public key...")
	keys, err := sshkeys.GetAll(doSvc)
	if err != nil {
		return nil, err
	}
	strKey := strings.Trim(string(pbkData), " \n\t")
	var publicKeyID int
	for _, key := range keys {
		if key.PublicKey == strKey {
			publicKeyID = key.ID
			break
		}
	}

	if publicKeyID == 0 {
		fmt.Println("Not found")
		createdKey, err := sshkeys.Create(doSvc, fmt.Sprintf("box-key-%v", strings.ToLower(projectName)), strKey)
		if err != nil {
			fmt.Println("Unable to post new SSH public key to DigitalOcean")
			return nil, err
		}

		config.PublicKeyID = createdKey.ID
	} else {
		fmt.Println("Found")
		config.PublicKeyID = publicKeyID
	}

	configBytes, err := yaml.Marshal(&config)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(configFilePath, configBytes, os.FileMode(0600))
	if err != nil {
		return nil, fmt.Errorf("Unable to write configuration file %v", configFilePath)
	}

	fmt.Println("All done!")

	return &config, nil
}

// Load loads the YAML configuration associated with the project name, or returns an error
func Load(projectName string) (*Config, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	configFilePath := filepath.Join(
		configDir,
		projectName,
		configFileName,
	)

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

// Save saves an existing configuration to disk
func (cfg *Config) Save() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}
	configFilePath := filepath.Join(
		configDir,
		cfg.ProjectName,
		configFileName,
	)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(configFilePath, data, os.FileMode(0600))
	return err
}

// DataDir returns the full path to the data directory which serves as the bind mount root
func (cfg *Config) DataDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	dataDir := filepath.Join(configDir, cfg.ProjectName, dataDirName)

	return dataDir, nil
}

// ProjectNameHash returns the first characters of the hex encoded SHA256 hash of the project name
func (cfg *Config) ProjectNameHash() string {
	if cfg.projNameHash == "" {
		hash := sha256.New()
		hash.Write([]byte(cfg.ProjectName))
		cfg.projNameHash = fmt.Sprintf("%x", hash.Sum(nil))[:12]
	}

	return cfg.projNameHash
}
