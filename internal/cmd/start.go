package cmd

import (
	"os"

	"github.com/owenrumney/lazytrivy/pkg/config"
	"github.com/owenrumney/lazytrivy/pkg/controllers/gui"
	"github.com/owenrumney/lazytrivy/pkg/logger"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

func startGUI(tab widgets.Tab) error {
	logger.Configure()

	workingDir, err := os.Getwd()
	if err != nil {
		return err
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
		return err
	}

	defer control.Close()

	// create the widgets
	if err := control.CreateWidgets(); err != nil {
		return err

	}

	// set up the initial view to be the images widget
	if err := control.Initialise(); err != nil {
		return err
	}

	// Enter the run loop - it's all in the gui from this point on
	return control.Run()
}
