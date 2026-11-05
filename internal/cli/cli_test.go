package cli

import (
	"fmt"
	"testing"

	// "github.com/spf13/cobra"
)


func TestRootCmd(t *testing.T) {
	if err := rootCmd.Execute(); err != nil {
		t.Error("rootCmd errored out ")
	} else {
		fmt.Println("passed with flying colors")
	}
}

