package config

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseMachineExample(t *testing.T) {
	got, err := Load("../../examples/machine.yaml")
	if err != nil {
		t.Fatalf("error parsing the config")
	}

	want := &Config{}

	want.APIVersion = APIVersion
	want.Kind = "Machine"
	want.Metadata = Metadata{Name: "web-01"}

	wantmachineSpec := MachineSpec{
		Provider: ProviderSpec{
			Type: "libvirt",
			URI:  "qemu:///system",
		},

		Image: ImageSpec{
			Catalog: "img-catalog/images.yaml",
			ID:      "ubuntu-24.04-lts-amd64-20261017",
		},

		Machine: Machine{
			VCPUS:    2,
			MemoryMb: 2048,
			DiskGb:   20,
		},

		Access: AccessSpec{
			SSHUser:        "ubuntu",
			PrivateKey:     "~/.ssh/id_ed25519",
			AuthorizedKeys: []string{"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIReplaceMeWithYourPublicKey operator@workstation"},
		},

		Network: NetworkSpec{Mode: "nat"},
	}
	want.Spec = wantmachineSpec

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parsed config mismatch:\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestParseServerExample(t *testing.T) {
	config, err := Load("../../examples/server.yaml")
	if err != nil {
		t.Fatalf("error parsing the config: %v", err)
	}

	nginxConfigContent := "server {\n" +
		"  listen 80 default_server;\n" +
		"  server_name _;\n" +
		"  root /var/www/html;\n" +
		"  index index.html;\n" +
		"}\n"

	want := &Config{
		APIVersion: APIVersion,
		Kind:       "Server",
		Metadata:   Metadata{Name: "web-01"},
		Spec: ServerSpec{
			Target: Target{
				Host: "192.168.122.45",
				Port: 22,
				User: "ubuntu",
				Auth: Auth{
					PrivateKey: "~/.ssh/id_ed25519",
				},
				HostKey: HostKey{
					Policy:     "strict",
					KnownHosts: "~/.ssh/known_hosts",
				},
				Become: Become{
					Method: "sudo",
					User:   "root",
				},
			},
			Resources: []Resources{
				{
					ID:                "admin-user",
					Type:              "user",
					Name:              "deploy",
					State:             "present",
					Groups:            []string{"sudo"},
					Shell:             "/bin/bash",
					Home:              "/home/deploy",
					SSHAuthorizedKeys: []string{"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIReplaceMeWithYourPublicKey operator@workstation"},
				},
				{
					ID:    "base-packages",
					Type:  "package",
					Names: []string{"nginx", "git"},
					State: "present",
				},
				{
					ID:           "nginx-config",
					Type:         "file",
					State:        "present",
					Owner:        "root",
					Group:        "root",
					Path:         "/etc/nginx/sites-available/default",
					Mode:         "0644",
					Content:      nginxConfigContent,
					Dependencies: []string{"base-packages"},
				},
				{
					ID:           "nginx-service",
					Type:         "service",
					Name:         "nginx",
					State:        "running",
					Enabled:      true,
					Dependencies: []string{"base-packages", "nginx-config"},
				},
			},
		},
	}

	if !reflect.DeepEqual(config, want) {
		t.Fatalf("parsed config mismatch:\ngot:  %#v\nwant: %#v", config, want)
	}
}

func TestUnknownField(t *testing.T) {
	input := `
apiVersion: kamui/v1
kind: Machine
metadata:
  name: web-01
spec:
  provider:
    type: libvirt
  image:
    catalog: img-catalog/images.yaml
    id: ubuntu-24.04-lts-amd64-20261017
  machine:
    vcpus: 2
    memory_mb: 2048
    disk_gb: 20
    unexpected: true
`

	_, err := Parse(strings.NewReader(input))
	if err == nil {
		t.Fatal("expected an error for an unknown field, got nil")
	}

	if !strings.Contains(err.Error(), "field unexpected not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnsupportedAPIVersion(t *testing.T) {
	input := `
apiVersion: kamui/v2
kind: Machine
metadata:
  name: web-01
spec:
  provider:
    type: libvirt
`

	_, err := Parse(strings.NewReader(input))
	if err == nil {
		t.Fatal("expected an error for an unsupported API version, got nil")
	}

	if !strings.Contains(err.Error(), "unsupported apiVersion") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnsupportedKind(t *testing.T) {
	input := `
apiVersion: kamui/v1
kind: Database
metadata:
  name: db-01
spec: {}
`

	_, err := Parse(strings.NewReader(input))
	if err == nil {
		t.Fatal("expected an error for an unsupported kind, got nil")
	}

	if !strings.Contains(err.Error(), "Database") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMalformedYAML(t *testing.T) {
	input := `
apiVersion: kamui/v1
kind: Machine
spec:
  provider: [
`

	_, err := Parse(strings.NewReader(input))
	if err == nil {
		t.Fatal("expected an error for malformed YAML, got nil")
	}
}
