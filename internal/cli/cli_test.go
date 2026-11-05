package cli

import (
	"bytes"
	"testing"
	"strings"

	// "github.com/spf13/cobra"
)

func executeCommand(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	cmd := NewRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)

	err := cmd.Execute()

	return strings.TrimSpace(buf.String()), err
}



func TestVersionCmd(t *testing.T) {
	out, err := executeCommand("version")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out != "version 0.1.0" {
		t.Errorf("got %q, want %q", out, "version 0.1.0")
	}
}

