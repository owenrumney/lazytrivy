package base

import (
	"github.com/awesome-gocui/gocui"
)

type Manager interface {
	AddViews(views ...gocui.Manager)
}
