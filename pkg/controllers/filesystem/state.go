package filesystem

import (
	"github.com/owenrumney/lazytrivy/pkg/output"
)

type state struct {
	workingDireectory string
	currentTarget     string
	currentReport     *output.Report
	currentResult     *output.Result
}
