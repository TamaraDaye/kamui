// Package config
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func Load(path string) {
}

func parseConfig(config *os.File) {
}

type specification interface {
	specs()
}

type Config struct {
	apiVersion string `yaml:"apiVersion"`
	configMeta metadata `yaml:"metadata"`
	kind       string `yaml:"kind"`
	spec       specification `yaml:"spec"`
}

type machineSpecification struct {
	provider machineProvisoner `yaml:"provider"`
	image    machineImage `yaml:"image"`
	machine  machineSpec `yaml:"machine"`
	access   machineAccess `yaml:"access"`
	network  networkCapabilites `yaml:"network"`
}

func (machineSpec *machineSpecification) spec() {}

type machineProvisoner struct {
	provider string `yaml:"provider"`
	uri      string  `yaml:"uri"`
}

type machineImage struct {
	catalog string `yaml:"catalog"`
	id      string `yaml:"id"`
}

type metadata struct {
	name string
}

type machineSpec struct {
	vcpus    uint `yaml:"vcpus"`
	memoryMb uint `yaml:"memory_mb"`
	diskGb   uint `yaml:"disk_gb"`
}

type machineAccess struct {
	sshUser        string `yaml:"ssh_user"`
	privateKey     string `yaml:"private_key"`
	authorizedKeys []string `yaml:"authorized_keys"`
}

type networkCapabilites struct {
	mode string `yaml:"mode"`
}

type ServerSpecification struct{}
