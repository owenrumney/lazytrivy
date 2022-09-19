package main

import "github.com/owenrumney/lazytrivy/internal/cmd"

func main() {

	rootCmd := cmd.GetRootCmd()

	_ = rootCmd.Execute()
}
