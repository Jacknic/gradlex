package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	versionCmd.Aliases = []string{"v"}
	rootCmd.AddCommand(versionCmd)
}

const versionName = "0.0.1"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Long:  `gradlex version https://github.com/Jacknic/gradlex `,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(versionName)
	},
}
