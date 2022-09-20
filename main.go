package main

import (
	"os"

	"github.com/owenrumney/lazytrivy/internal/cmd"
)

func main() {

	// if no args are passed, open in image mode
	if len(os.Args[1:]) == 0 {
		os.Args = append(os.Args, "image")
	}

	rootCmd := cmd.GetRootCmd()
	_ = rootCmd.Execute()
}
