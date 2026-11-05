package config

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

func decodeKind[T any](data []byte) (*Config, error) {
	var doc document[T]

	if err := decodeStrict(data, &doc); err != nil {
		return nil, err
	}

	return &Config {
		APIVersion: doc.APIVersion,
		Kind: doc.Kind, 
		Metadata: doc.Metadata,
		Spec: doc.Spec,
	}, nil
}


func Load(path string) (*Config,  error)  {
	f, err := os.Open(path)

	if err != nil {
		return nil, fmt.Errorf("couldn't open provided config %w", err)
	}

	defer f.Close()

	cfg, err := Parse(f)

	if err != nil {
		return nil, fmt.Errorf("could not load config %w", err)
	}

	fmt.Println(cfg)
	return cfg, nil
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

	switch header.Kind {
	case "Machine":
		return decodeKind[MachineSpec](data)
	case "Server":
		return decodeKind[ServerSpec](data)
	default: 
		return nil, fmt.Errorf("invalid machine kind specified %q", header.Kind)

	}
}

func decodeStrict(data []byte, out any) error {
	r := bytes.NewReader(data)

	decoder := yaml.NewDecoder(r)

	decoder.KnownFields(true)


	if err := decoder.Decode(out); err != nil {
		return err
	}
	return nil 
}

