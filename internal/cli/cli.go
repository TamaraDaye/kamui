// Package cli
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kamui",
	Short: "A simple vm provisioner CLI tool",
	Long:  " ",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "display version for cli",
	Long:  "This command displays the version for kamui",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("version 0.1.0")
	},
}

var helpCmd = &cobra.Command{
	Use:   "help",
	Short: "help information on any command",
	Long:  "help information on any command [command]",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
        fmt.Println(args[0])
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
	fmt.Println("test build")
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(helpCmd)
}
