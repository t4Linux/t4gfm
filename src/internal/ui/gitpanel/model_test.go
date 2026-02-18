package gitpanel

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestTabSwitching(t *testing.T) {
	m := New()
	assert.Equal(t, "git", m.ActiveTabName())
	m.NextTab()
	assert.Equal(t, "clipboard", m.ActiveTabName())
	m.PrevTab()
	assert.Equal(t, "git", m.ActiveTabName())
}

func TestClipboardTabRender(t *testing.T) {
	m := New()
	m.SetDimensions(50, 8)
	m.SetClipboard([]string{"/tmp/a.txt"}, false)
	m.NextTab()
	out := ansi.Strip(m.Render(true))
	assert.Contains(t, out, "G.. | Clipboard")
	assert.Contains(t, out, "mode: copy")
}

func TestClipboardTabScrolling(t *testing.T) {
	tmp := t.TempDir()
	paths := []string{
		filepath.Join(tmp, "a.txt"),
		filepath.Join(tmp, "b.txt"),
		filepath.Join(tmp, "c.txt"),
		filepath.Join(tmp, "d.txt"),
	}
	for _, p := range paths {
		assert.NoError(t, os.WriteFile(p, []byte("x"), 0o644))
	}

	m := New()
	m.SetDimensions(50, 8)
	m.SetClipboard(paths, false)
	m.NextTab()

	out := ansi.Strip(m.Render(true))
	assert.Contains(t, out, "1/4")

	m.ListDown()
	out = ansi.Strip(m.Render(true))
	assert.Contains(t, out, "2/4")

	m.ListUp()
	out = ansi.Strip(m.Render(true))
	assert.Contains(t, out, "1/4")
}

func TestClipboardTabScrollingSmallHeight(t *testing.T) {
	tmp := t.TempDir()
	paths := []string{
		filepath.Join(tmp, "a.txt"),
		filepath.Join(tmp, "b.txt"),
		filepath.Join(tmp, "c.txt"),
	}
	for _, p := range paths {
		assert.NoError(t, os.WriteFile(p, []byte("x"), 0o644))
	}

	m := New()
	m.SetDimensions(20, 5)
	m.SetClipboard(paths, false)
	m.NextTab()

	out := ansi.Strip(m.Render(true))
	assert.Contains(t, out, "a.txt")

	m.ListDown()
	out = ansi.Strip(m.Render(true))
	assert.Contains(t, out, "b.txt")
}
