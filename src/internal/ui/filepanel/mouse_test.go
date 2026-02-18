package filepanel

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/t4Linux/t4gfm/src/internal/common"
)

func TestClickedItemIndexWithoutSearchBar(t *testing.T) {
	oldExtraCols := common.Config.FilePanelExtraColumns
	common.Config.FilePanelExtraColumns = 1
	t.Cleanup(func() { common.Config.FilePanelExtraColumns = oldExtraCols })

	panel := Model{SearchBar: textinput.New()}
	panel.columns = []columnDefinition{{Name: "Name"}, {Name: "Size"}}
	panel.element = []Element{{Name: "a"}, {Name: "b"}, {Name: "c"}}

	if got := panel.ClickedItemIndex(4); got != 0 {
		t.Fatalf("expected first row to map to item 0, got %d", got)
	}
	if got := panel.ClickedItemIndex(5); got != 1 {
		t.Fatalf("expected second row to map to item 1, got %d", got)
	}
}

func TestClickedItemIndexWithSearchBar(t *testing.T) {
	oldExtraCols := common.Config.FilePanelExtraColumns
	common.Config.FilePanelExtraColumns = 1
	t.Cleanup(func() { common.Config.FilePanelExtraColumns = oldExtraCols })

	panel := Model{SearchBar: textinput.New()}
	panel.columns = []columnDefinition{{Name: "Name"}, {Name: "Size"}}
	panel.element = []Element{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	panel.SearchBar.SetValue("x")

	if got := panel.ClickedItemIndex(5); got != 0 {
		t.Fatalf("expected first row with search to map to item 0, got %d", got)
	}
	if got := panel.ClickedItemIndex(6); got != 1 {
		t.Fatalf("expected second row with search to map to item 1, got %d", got)
	}
}
