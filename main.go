package main

import (
	"fmt"
	"os"

	"github.com/owenrumney/lazytrivy/pkg/controllers/gui"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

func main() {

	if len(os.Args) == 1 {
		if err := startGUI(widgets.VulnerabilitiesTab, ""); err != nil {
			fail(err)
		}
		return
	} else if os.Args[1] == "aws" {
		if err := startGUI(widgets.AWSTab, ""); err != nil {
			fail(err)
		}
		return
	} else if os.Args[1] == "fs" {
		cwd, err := os.Getwd()
		if err != nil {
			fail(err)
		}
		if len(os.Args) == 3 {
			if _, err := os.Stat(os.Args[2]); err == nil {
				cwd = os.Args[2]
			}
		}
		if err := startGUI(widgets.FileSystemTab, cwd); err != nil {
			fail(err)
		}
		return
	} else {
		if err := startGUI(widgets.VulnerabilitiesTab, ""); err != nil {
			fail(err)
		}
		return
	}

}

func fail(err error) {
	_, _ = fmt.Fprintf(os.Stderr, "Error: %s", err)
	os.Exit(1)
}

func startGUI(tab widgets.Tab, workingDir string) error {
	control, err := gui.New(tab, workingDir)
	if err != nil {
		fail(err)

	}

	defer control.Close()

	// create the widgets
	if err := control.CreateWidgets(); err != nil {
		fail(err)

	}

	// set up the initial view to be the images widget
	if err := control.Initialise(); err != nil {
		return err
	}

	if control.IsDockerDesktop() {
		control.ShowDockerDesktopWarning()
	}

	// Enter the run loop - it's all in the gui from this point on
	if err := control.Run(); err != nil {
		fail(err)

	}

	return nil
}
