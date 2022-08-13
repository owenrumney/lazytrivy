package main

import (
	"fmt"
	"os"

	"github.com/owenrumney/lazytrivy/pkg/gui"
)

func main() {
	g, err := gui.New()
	if err != nil {
		fail(err)
	}

	// create the widgets
	if err := g.CreateWidgets(); err != nil {
		fail(err)
	}

	// set up the initial view to be the images widget
	g.Initialise()

	// Enter the run loop - its all in the gui from this point on
	if err := g.Run(); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Printf("Error: %s", err)
	os.Exit(1)
}
