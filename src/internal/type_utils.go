package internal

import (
	"github.com/t4Linux/t4gfm/src/internal/common"
)

// ================ String method for easy logging =====================

func (f focusPanelType) String() string {
	switch f {
	case nonePanelFocus:
		return "nonePanelFocus"
	case processBarFocus:
		return "processBarFocus"
	case sidebarFocus:
		return "sidebarFocus"
	case metadataFocus:
		return "metadataFocus"
	case gitPanelFocus:
		return "gitPanelFocus"
	default:
		return common.InvalidTypeString
	}
}
