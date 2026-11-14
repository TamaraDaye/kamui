package config

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

func decodeKind[T any](data []byte) (*Config, error) {
	var configDoc configDocument[T]

	if err := decodeStrict(data, &configDoc); err != nil {
		return nil, err
	}

	return &Config {
		APIVersion: configDoc.APIVersion, 
		Kind: configDoc.Kind,
		Metadata: configDoc.Metadata,
		Spec: &configDoc.Spec,
	}, nil

}

func Validate(configs []string) ([]*Config, error) {
	validConfigs := make([]*Config, len(configs))


	for _, file := range configs {
		config, err := Load(file)

		if err != nil {
			return nil, fmt.Errorf("invalid config specified %w", err)
		}

		validConfigs = append(validConfigs, config)
	}

	return validConfigs, nil
	
}

func Load(path string) (*Config, error) {
	f, err := os.Open(path)

	if err != nil {
		return nil, fmt.Errorf("couldn't open provided config %w", err)
	}

	defer f.Close()

	cfg, err := parse(f)

	if err != nil {
		return nil, fmt.Errorf("could not load config %w", err)
	}

	return cfg, nil
}

func parse(r io.Reader) (*Config, error) {
	data, err := io.ReadAll(r)

	if err != nil {
		return nil, fmt.Errorf("couldn't parse provided config %w", err)
	}

	var header struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
	}

	if err := yaml.Unmarshal(data, &header); err != nil {
		return nil, fmt.Errorf("error parsing config %w", err)
	}

	if header.APIVersion != APIVersion {
		return nil, fmt.Errorf("unsupported apiVersion %q", header.APIVersion)
	}


	if header.Kind != "Machine" || header.Kind != "Server" {
		return nil, fmt.Errorf("unsupported provisioning kind %q", header.Kind)
	}


	var (
		config *Config 
	)


	switch header.Kind {
	case "Machine":
		config, err = decodeKind[MachineSpec](data)
	case "Server":
		config, err = decodeKind[ServerSpec](data)
	default:
		return nil, fmt.Errorf("invalid machine kind specified %q", header.Kind)

	}

	if err != nil {
		return nil, fmt.Errorf("could not parse config %w", err)
	}


	switch config.Spec.(type) {

	case *MachineSpec:
		spec := config.Spec.(*MachineSpec)
		validateMachineSpec(spec)

	case *ServerSpec: 
		spec := config.Spec.(*ServerSpec)
		validateServerSpec(spec)
	}


	return config, nil

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



func validateServerSpec(spec *ServerSpec) error {

	id := make(map[string]struct{})

	for _, resource := range spec.Resources {
		if resource.ID == "" {
			return fmt.Errorf("unspecified resource id(s)")
		}
		if _, seen := id[resource.ID]; seen {
			return fmt.Errorf("resource specified twice confliciting (ids)")
		} 



		id[resource.ID] = struct{}{}
	}
	return nil
}


func validateMachineSpec(spec *MachineSpec) error {


	return nil
}
