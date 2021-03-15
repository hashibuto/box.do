package manifest

type Build struct {
	Dockerfile string `yaml:"dockerfile"`
	Context    string `yaml:"context"`
}

type Path struct {
	Pattern string `yaml:"pattern"`
	Type    string `yaml:"type"`
}

type Routing struct {
	Path Path `yaml:"path"`
	Port int  `yaml:"port"`
}

type Service struct {
	ContainerName string            `yaml:"container_name"`
	Hostname      string            `yaml:"hostname"`
	Routing       Routing           `yaml:"routing"`
	Environment   map[string]string `yaml:"environment"`
	Volumes       []string          `yaml:"volumes"`
	Image         string            `yaml:"image"`
	DependsOn     string            `yaml:"depends_on"`
}

type Manifest struct {
	Project  string             `yaml:"project"`
	Services map[string]Service `yaml:"services"`
}
