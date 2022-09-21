package main

import (
	"os"

	"github.com/owenrumney/lazytrivy/internal/cmd"
	"github.com/owenrumney/lazytrivy/pkg/logger"
)

func main() {
	// configure the logger
	logger.Configure()

	// if no args are passed, open in image mode
	if len(os.Args[1:]) == 0 {
		logger.Infof("No arguments passed, opening in image mode")
		os.Args = append(os.Args, "image")
	}

	rootCmd := cmd.GetRootCmd()
	if err := rootCmd.Execute(); err != nil {
		logger.Errorf("Error executing command: %v", err)
		os.Exit(1)
	}
}
