// Package config
package config

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

const APIVersion = "kamui/v1"

type specification interface {
	specs()
}

type machineDocument struct {
	APIVersion string      `yaml:"apiVersion"`
	Kind       string      `yaml:"kind"`
	Metadata   Metadata    `yaml:"metadata"`
	Spec       MachineSpec `yaml:"spec"`
}

type Config struct {
	APIVersion string        `yaml:"apiVersion"`
	Metadata   Metadata      `yaml:"metadata"`
	Kind       string        `yaml:"kind"`
	Spec       any `yaml:"spec"`
}

type MachineSpec struct {
	Provider ProviderSpec `yaml:"provider"`
	Image    ImageSpec    `yaml:"image"`
	Machine  Machine      `yaml:"machine"`
	Access   AccessSpec   `yaml:"access"`
	Network  networkSpec  `yaml:"network"`
}

func (machineSpec *MachineSpec) spec() {}

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

type networkSpec struct {
	mode string `yaml:"mode"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}

	cfg, err := Parse(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("parse config %q: %w", path, err)
	}

	return cfg, nil
}

func Parse(r io.Reader) (*Config, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var header struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
	}

	if err := yaml.Unmarshal(data, &header); err != nil {
		return nil, err
	}

	if header.APIVersion != APIVersion {
		return nil, fmt.Errorf("unsupported apiVersion %q", header.APIVersion)
	}

	switch header.Kind {
	case "Machine":
		var doc machineDocument

		if err := decodeStrict(data, &doc); err != nil {
			return nil, err
		}

		return &Config{
			APIVersion: doc.APIVersion,
			Kind:       doc.Kind,
			Metadata:   doc.Metadata,
			Spec:       &doc.Spec,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported or not-yet-implemented kind %q", header.Kind)
	}
}


func decodeStrict(data []byte, out any) error {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)

	if err := dec.Decode(out); err != nil {
		return err
	}

	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple YAML documents are not supported")
		}
		return err
	}

	return nil
}

type ServerSpecification struct{}
