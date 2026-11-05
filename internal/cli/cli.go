// Package cli
package cli

import (
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "kamui",
		Short: "A simple vm provisioner CLI tool",
		Long:  " ",
	}

	rootCmd.AddCommand(newVersionCmd())

	return rootCmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "display version for cli",
		Long:  "This command displays the version for kamui",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("version 0.1.0")
		},
	}
}


func Execute() {
	cobra.CheckErr(NewRootCmd().Execute())
}
