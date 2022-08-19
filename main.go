package main

import (
	"fmt"
	"os"

	"github.com/owenrumney/lazytrivy/pkg/controllers/gui"
)

func main() {
	control, err := gui.New()
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
		fail(err)
	}

	// Enter the run loop - it's all in the gui from this point on
	if err := control.Run(); err != nil {
		fail(err)
	}
}

func fail(err error) {
	_, _ = fmt.Fprintf(os.Stderr, "Error: %s", err)
	os.Exit(1)
}
