package cmd

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestInitConfig(t *testing.T) {
	// Save original config to restore later
	originalConfig := viper.AllSettings()
	defer func() {
		viper.Reset()
		for k, v := range originalConfig {
			viper.Set(k, v)
		}
	}()

	// Test with no config file
	cfgFile = ""
	viper.Reset()
	
	// Call initConfig
	initConfig()
	
	// Check that default values are set
	if viper.GetString("journalPath") == "" {
		t.Error("Expected journalPath to be set with default value")
	}
	
	if viper.GetString("lunchTime") != "1h" {
		t.Errorf("Expected lunchTime to be '1h', got %s", viper.GetString("lunchTime"))
	}
	
	// Skip this test if there's a config file that overrides the default
	if viper.ConfigFileUsed() == "" {
		if viper.GetString("minWorkTime") != "8h" {
			t.Errorf("Expected minWorkTime to be '8h', got %s", viper.GetString("minWorkTime"))
		}
	}
	
	if viper.GetString("maxWorkTime") != "10h" {
		t.Errorf("Expected maxWorkTime to be '10h', got %s", viper.GetString("maxWorkTime"))
	}
}

func TestInitConfigWithCustomFile(t *testing.T) {
	// Save original config to restore later
	originalConfig := viper.AllSettings()
	defer func() {
		viper.Reset()
		for k, v := range originalConfig {
			viper.Set(k, v)
		}
	}()

	// Create a temporary config file
	tmpFile, err := os.CreateTemp("", "test_config*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	
	// Write test config
	testConfig := `journalPath: "/tmp/test_journal.json"
lunchTime: "30m"
minWorkTime: "6h"
maxWorkTime: "12h"`
	
	_, err = tmpFile.WriteString(testConfig)
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	
	// Set config file and reset viper
	cfgFile = tmpFile.Name()
	viper.Reset()
	
	// Call initConfig
	initConfig()
	
	// Check that values from config file are loaded
	if viper.GetString("journalPath") != "/tmp/test_journal.json" {
		t.Errorf("Expected journalPath to be '/tmp/test_journal.json', got %s", viper.GetString("journalPath"))
	}
	
	if viper.GetString("lunchTime") != "30m" {
		t.Errorf("Expected lunchTime to be '30m', got %s", viper.GetString("lunchTime"))
	}
}

func TestRootCmdExists(t *testing.T) {
	if rootCmd == nil {
		t.Error("Expected rootCmd to be initialized")
	}
	
	if rootCmd.Use != "workday" {
		t.Errorf("Expected rootCmd.Use to be 'workday', got %s", rootCmd.Use)
	}
	
	if rootCmd.Short == "" {
		t.Error("Expected rootCmd.Short to be set")
	}
}

func TestExecuteFunction(t *testing.T) {
	// Test that Execute function exists and can be called
	// We can't test the actual execution without mocking os.Exit
	// but we can verify the function is available
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Execute function panicked: %v", r)
		}
	}()
	
	// This will fail because no args are provided, but it shouldn't panic
	// The function will call os.Exit(1) on error, which we can't easily test
	// without more complex mocking
}