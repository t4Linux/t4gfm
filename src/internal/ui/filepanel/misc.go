package filepanel

import "github.com/t4Linux/t4gfm/src/internal/common"

func (p PanelMode) String() string {
	switch p {
	case SelectMode:
		return "selectMode"
	case BrowserMode:
		return "browserMode"
	default:
		return common.InvalidTypeString
	}
}
