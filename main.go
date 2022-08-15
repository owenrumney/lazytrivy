package main

import (
	"fmt"
	"os"

	"github.com/owenrumney/lazytrivy/pkg/controller"
)

func main() {
	c, err := controller.New()
	if err != nil {
		fail(err)
	}

	defer c.Close()

	// create the widgets
	if err := c.CreateWidgets(); err != nil {
		fail(err)
	}

	// set up the initial view to be the images widget
	c.Initialise()

	// Enter the run loop - its all in the gui from this point on
	if err := c.Run(); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Printf("Error: %s", err)
	os.Exit(1)
}
