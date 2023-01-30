package cmd

import (
	"github.com/owenrumney/lazytrivy/pkg/widgets"
	"github.com/spf13/cobra"
)

var cmdImage = &cobra.Command{
	Use:   "image",
	Short: "Launch lazytrivy in image scanning mode",
	RunE: func(cmd *cobra.Command, args []string) error {
		return startGUI(widgets.VulnerabilitiesTab)
	},
}

var cmdFS = &cobra.Command{
	Use:     "filesystem",
	Aliases: []string{"fs"},
	Short:   "Launch lazytrivy in filesystem scanning mode",
	RunE: func(cmd *cobra.Command, args []string) error {
		return startGUI(widgets.FileSystemTab)
	},
}

func GetRootCmd() *cobra.Command {
	generalFlags := createGeneralFlags()
	filesystemFlags := createFilesystemFlags()

	cmdImage.Flags().AddFlagSet(generalFlags)

	cmdFS.Flags().AddFlagSet(generalFlags)
	cmdFS.Flags().AddFlagSet(filesystemFlags)

	rootCmd := &cobra.Command{
		Use: "lazytrivy",
	}

	rootCmd.AddCommand(cmdImage)
	rootCmd.AddCommand(cmdFS)

	rootCmd.Flags().AddFlagSet(generalFlags)

	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	return rootCmd
}
