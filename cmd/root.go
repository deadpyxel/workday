package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// Version information
var (
	appVersion = "dev"
	appCommit  = "none"
	appDate    = "unknown"
	appBuiltBy = "unknown"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "workday",
	Short: "A CLI tool for managing and reporting workday entries",
	Long: `Workday is a CLI tool that allows you to manage your workday entries and generate reports.

Workday provides a set of commands and subcommands that you can use to add, update, and report workday entries.
Workday works like a journal where you can add notes to you day, and track the start and end of the day.

For example, you can use 'workday start' to start a new workday, 'workday note' to add notes to the current day.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	Example: "workday start",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
// SetVersionInfo sets the version information from main
func SetVersionInfo(version, commit, date, builtBy string) {
	appVersion = version
	appCommit = commit
	appDate = date
	appBuiltBy = builtBy
	rootCmd.Version = version
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.workday.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".workday" (without extension).
		viper.AddConfigPath(home + "/.config/workday")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config.yaml")
	}

	// Set default value for journalPath. Will use your HOME if not set.
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)
	viper.SetDefault("journalPath", home+"/journal.json")
	viper.SetDefault("lunchTime", "1h")
	viper.SetDefault("minWorkTime", "8h")
	viper.SetDefault("maxWorkTime", "10h")

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
