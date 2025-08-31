package main

import "license-audit/cmd"

// Version information - set during build
var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	cmd.SetVersion(version, buildTime)
	cmd.Execute()
}