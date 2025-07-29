package main

import (
	"github.com/deadpyxel/workday/cmd"
)

// Version information set by goreleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	cmd.SetVersionInfo(version, commit, date, builtBy)
	cmd.Execute()
}
