package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	envCmd.Aliases = []string{"e"}
	rootCmd.AddCommand(envCmd)
}

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Print env or default",
	Long:  ` `,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("GRADLE_HOME: ", getGradleHome())
		fmt.Println("GRADLE_USER_HOME: ", getGradleUserHome())
		fmt.Println("GRADLE_PROXY: ", getGradleDistProxy())
	},
}
