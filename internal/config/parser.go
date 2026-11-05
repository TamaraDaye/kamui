package config

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

func Load(path string) error {
	data, err := os.ReadFile(path)

	if err != nil {
		return fmt.Errorf("couldn't open provided config %w", err)
	}

	cfg, err := Parse(bytes.NewReader(data))


	if err != nil {
		return fmt.Errorf("could not load config %w", err)
	}

	fmt.Println(cfg)
	return nil
}


func Parse(r io.Reader) (*Config, error ) {
	data, err := io.ReadAll(r)

	if err != nil {
		return nil, fmt.Errorf("couldn't parse provided config %w", err)
	}

	var header struct {
		APIVersion string `yaml:"apiVersion"`
		Kind string `yaml:"kind"`
	}

	if err := yaml.Unmarshal(data, &header); err != nil {
		return nil, fmt.Errorf("error parsing config %w", err)
	}

	if header.APIVersion != APIVersion {
		return nil, fmt.Errorf("unsupported apiVersion %q", header.APIVersion)
	}

	cfg := &Config{}


	switch header.Kind {
	case "Machine":
		cfg.Spec = &MachineSpec{}
	default: 
		return nil, fmt.Errorf("invalid header kind %q", header.Kind)

	}

	if err := decodeStrict(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func decodeStrict(data []byte, cfg *Config) error {
	r := bytes.NewReader(data)


	decoder := yaml.NewDecoder(r)

	decoder.KnownFields(true)


	if err := decoder.Decode(cfg); err != nil {
		return err
	}
	return nil 
}


