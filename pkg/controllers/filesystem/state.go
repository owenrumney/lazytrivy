package filesystem

import (
	"context"

	"github.com/owenrumney/lazytrivy/pkg/output"
)

type state struct {
	workingDireectory string
	currentTarget     string
	currentReport     *output.Report
	currentResult     *output.Result
}

func (s *state) runVulnerabilityScan(ctx context.Context, imageName string) error {

	return nil
}
