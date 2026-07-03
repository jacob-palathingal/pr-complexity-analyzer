package cmd

import "fmt"

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func versionString() string {
	return fmt.Sprintf("%s commit=%s date=%s", version, commit, date)
}
