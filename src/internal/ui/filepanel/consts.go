package filepanel

import (
	"time"

	"github.com/t4Linux/t4gfm/src/internal/common"
)

const (
	contentPadding = 3
	MinHeight      = contentPadding + common.BorderPadding + 1
	MinWidth       = 18 // minimal width for rename input to render

	FileSizeColumnWidth       = 15
	ModifyTimeSizeColumnWidth = 18
	PermissionsColumnWidth    = 12
	ColumnHeaderHeight        = 1

	// Delimiter between columns in the file panel.
	ColumnDelimiter      = "  "
	ReRenderChunkDivisor = 100
	ReRenderMaxDelay     = 3

	nonFocussedPanelReRenderTime = 3 * time.Second
	gitStatusRefreshInterval     = 2 * time.Second
)
