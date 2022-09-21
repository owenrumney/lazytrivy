package cmd

import (
	"github.com/spf13/pflag"
)

var (
	// general flags
	dockerHost   *string
	debugEnabled *bool
	traceEnabled *bool

	// filesystem flags
	scanPath string
)

func createGeneralFlags() *pflag.FlagSet {
	generalFlags := pflag.NewFlagSet("general", pflag.ExitOnError)

	dockerHost = generalFlags.String("docker-host", "unix:///var/run/docker.sock", "Docker host to connect to")
	debugEnabled = generalFlags.Bool("debug", false, "Launch with debug logging")
	traceEnabled = generalFlags.Bool("trace", false, "Launch with trace logging")

	return generalFlags
}

func createFilesystemFlags() *pflag.FlagSet {
	filesystemFlags := pflag.NewFlagSet("filesystem", pflag.ExitOnError)
	filesystemFlags.StringVar(&scanPath, "path", "", "Path to scan")
	return filesystemFlags
}
