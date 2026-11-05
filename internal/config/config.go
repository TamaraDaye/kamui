// Package config
package config

import "net"

const APIVersion = "kamui/v1"

type Config struct {
	APIVersion string   `yaml:"apiVersion"`
	Metadata   Metadata `yaml:"metadata"`
	Kind       string   `yaml:"kind"`
	Spec       any      `yaml:"spec"`
}

type MachineSpec struct {
	Provider ProviderSpec `yaml:"provider"`
	Image    ImageSpec    `yaml:"image"`
	Machine  Machine      `yaml:"machine"`
	Access   AccessSpec   `yaml:"access"`
	Network  NetworkSpec  `yaml:"network"`
}


type ProviderSpec struct {
	Type string `yaml:"provider"`
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
	mode string `yaml:"mode"`
}


type ServerSpec struct {
	target Target
}


type Target struct {
	Host  net.Addr `yaml:"host"`
	Port int `yaml:"port"`
	User string `yaml:"user"`
}

type Resources struct {
	ID   string `yaml:"id"`
	Type string `yaml:"type"`
	Name string `yaml:"names,omitempty"`
	Names []string `yaml:"names,omitempty"`
	State string `yaml:"state"`
	Groups []string `yaml:"groups"`
	Group string `yaml:"group"`
	Shell string `yaml:"shell"`
}
