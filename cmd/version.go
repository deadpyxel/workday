package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Print detailed version information including version, commit, build date and Go version.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("workday version %s\n", appVersion)
		fmt.Printf("commit: %s\n", appCommit)
		fmt.Printf("built: %s\n", appDate)
		fmt.Printf("built by: %s\n", appBuiltBy)
		fmt.Printf("go version: %s\n", runtime.Version())
		fmt.Printf("platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}