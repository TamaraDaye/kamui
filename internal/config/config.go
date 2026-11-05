// Package config
package config

const APIVersion = "kamui/v1"

type Config struct {
	APIVersion string   `yaml:"apiVersion"`
	Metadata   Metadata `yaml:"metadata"`
	Kind       string   `yaml:"kind"`
	Spec       any      `yaml:"spec"`
}

type document[T any] struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       T        `yaml:"spec"`
}

type MachineSpec struct {
	Provider ProviderSpec `yaml:"provider"`
	Image    ImageSpec    `yaml:"image"`
	Machine  Machine      `yaml:"machine"`
	Access   AccessSpec   `yaml:"access"`
	Network  NetworkSpec  `yaml:"network"`
}

type ProviderSpec struct {
	Type string `yaml:"type"`
	URI  string `yaml:"uri"`
}

type ImageSpec struct {
	Catalog string `yaml:"catalog"`
	ID      string `yaml:"id"`
}

type Metadata struct {
	Name string
}

type Machine struct {
	VCPUS    uint `yaml:"vcpus"`
	MemoryMb uint `yaml:"memory_mb"`
	DiskGb   uint `yaml:"disk_gb"`
}

type AccessSpec struct {
	SSHUser        string   `yaml:"ssh_user"`
	PrivateKey     string   `yaml:"private_key"`
	AuthorizedKeys []string `yaml:"authorized_keys"`
}

type NetworkSpec struct {
	Mode string `yaml:"mode"`
}

type Resources struct {
	ID                string   `yaml:"id"`
	Type              string   `yaml:"type"`
	Name              string   `yaml:"name,omitempty"`
	Names             []string `yaml:"names,omitempty"`
	State             string   `yaml:"state"`
	Groups            []string `yaml:"groups,omitempty"`
	Group             string   `yaml:"group,omitempty"`
	Shell             string   `yaml:"shell,omitempty"`
	Home              string   `yaml:"home,omitempty"`
	SSHAuthorizedKeys []string `yaml:"ssh_authorized_keys"`
	Owner             string   `yaml:"owner"`
	Dependencies      []string `yaml:"depends_on"`
	Enabled           bool     `yaml:"enabled"`
	Path              string   `yaml:"path"`
	Mode              string   `yaml:"mode"`
	Content           string   `yaml:"content"`
}

type Target struct {
	Host    string  `yaml:"host"`
	Port    int     `yaml:"port"`
	User    string  `yaml:"user"`
	Auth    Auth    `yaml:"auth"`
	HostKey HostKey `yaml:"host_key"`
	Become  Become  `yaml:"become"`
}

type Auth struct {
	PrivateKey string `yaml:"private_key"`
}

type HostKey struct {
	Policy     string `yaml:"policy"`
	KnownHosts string `yaml:"known_hosts"`
}

type Become struct {
	Method string `yaml:"method"`
	User   string `yaml:"user"`
}

type ServerSpec struct {
	Target    Target      `yaml:"target"`
	Resources []Resources `yaml:"resources"`
}
