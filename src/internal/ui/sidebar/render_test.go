package sidebar

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestRenderSectionPanelClampsStartWhenRenderIndexOutsideSection(t *testing.T) {
	s := defaultTestModel(0, 100, 20, 5, 4, 2)
	s.width = 40

	sections := s.sectionBuckets()
	out := ansi.Strip(s.renderSectionPanel("Files", 8, true, sections.list, "", true, true))

	assert.Contains(t, out, "Dir")
}
