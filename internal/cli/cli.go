// Package cli
package cli

import (
	"github.com/spf13/cobra"
	"github.com/TamaraDaye/kamui/internal/config"
)

func NewRootCmd(cmds ...func()*cobra.Command) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "kamui",
		Short: "A simple vm provisioner CLI tool",
		Long:  " ",
	}

	rootCmd.AddCommand(newVersionCmd())


	for _, cmd := range cmds {
		rootCmd.AddCommand(cmd())
	}

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
	cobra.CheckErr(NewRootCmd(newValidateCmd, newVersionCmd).Execute())
}


func newValidateCmd() *cobra.Command {
	return &cobra.Command {
		Use: "validate", 
		Short: "validate config provider by user", 
		Long : "This Command validates a config(s) provided by user", 
		Args: func(cmd *cobra.Command, args []string) error { 
			if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
				return err
			}

			return nil 
		},
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
}
