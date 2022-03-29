package version

import "fmt"

const Version = "v0.1.4"

// getVersion Compulsory minimum version, Minimum downward compatibility to this version
func getVersion() string {
	return Version
}

// PrintVersion print currently version info
func PrintVersion() {
	fmt.Printf("Version: %s\nCore version: %s\nSame core version of generate-tool\n", Version, getVersion())
}
