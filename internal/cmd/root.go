package cmd

import (
	"github.com/owenrumney/lazytrivy/pkg/widgets"
	"github.com/spf13/cobra"
)

var imageCommand = &cobra.Command{
	Use:   "image",
	Short: "Launch lazytrivy in image scanning mode",
	RunE: func(cmd *cobra.Command, args []string) error {
		return start(widgets.ImagesTab)
	},
}

var fsCommand = &cobra.Command{
	Use:     "filesystem",
	Aliases: []string{"fs"},
	Short:   "Launch lazytrivy in filesystem scanning mode",
	RunE: func(cmd *cobra.Command, args []string) error {
		return start(widgets.FileSystemTab)
	},
}

var k8sCommand = &cobra.Command{
	Use:     "kubernetes",
	Aliases: []string{"k8s"},
	Short:   "Launch lazytrivy in kubernetes scanning mode",
	RunE: func(cmd *cobra.Command, args []string) error {
		return start(widgets.K8sTab)
	},
}

func GetRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "lazytrivy",
	}

	generalFlags := createGeneralFlags()
	filesystemFlags := createFilesystemFlags()

	rootCmd.AddCommand(imageCommand)
	rootCmd.AddCommand(fsCommand)
	rootCmd.AddCommand(k8sCommand)

	imageCommand.Flags().AddFlagSet(generalFlags)
	fsCommand.Flags().AddFlagSet(generalFlags)
	fsCommand.Flags().AddFlagSet(filesystemFlags)
	k8sCommand.Flags().AddFlagSet(generalFlags)
	rootCmd.Flags().AddFlagSet(generalFlags)
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	return rootCmd
}
