package cmd

import (
	"fmt"
	"os"

	"github.com/owenrumney/lazytrivy/pkg/config"
	"github.com/owenrumney/lazytrivy/pkg/controllers/gui"
	"github.com/owenrumney/lazytrivy/pkg/logger"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

func startGUI(tab widgets.Tab) error {
	workingDir, err := os.Getwd()
	if err != nil {
		fail(err)
	}

	cfg, err := config.Load()
	if err != nil {
		return err

	}
	if cfg.Debug || *debugEnabled {
		logger.EnableDebugging()
	}
	if cfg.Trace || *traceEnabled {
		logger.EnableTracing()
	}

	if *dockerHost != "" {
		cfg.DockerEndpoint = *dockerHost
	}

	if scanPath != "" {
		cfg.Filesystem.WorkingDirectory = scanPath
	} else {
		cfg.Filesystem.WorkingDirectory = workingDir
	}

	control, err := gui.New(tab, cfg)
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

func fail(err error) {
	_, _ = fmt.Fprintf(os.Stderr, "Error: %s", err)
	os.Exit(1)
}
