package filepanel

import (
	"os"

	"github.com/charmbracelet/x/ansi"

	"github.com/t4Linux/t4gfm/src/internal/common"
)

func (m *Model) IsFocusedNameTruncated() bool {
	if m.EmptyOrInvalid() || len(m.columns) == 0 {
		return false
	}

	elem := m.GetFocusedItem()
	isLink := false
	if elem.Info != nil {
		isLink = elem.Info.Mode()&os.ModeSymlink != 0
	}

	style := common.GetElementIcon(elem.Name, elem.Directory, isLink, common.Config.Nerdfont)
	iconData := style.Icon + " "
	if isLink {
		iconData = style.Icon + " -> "
	}

	nameColumnWidth := m.columns[0].Size
	prefixWidth := ansi.StringWidth("  ")
	filenameWidth := nameColumnWidth - prefixWidth - ansi.StringWidth(iconData)
	if filenameWidth <= 0 {
		return true
	}

	return ansi.StringWidth(elem.Name) > filenameWidth
}
