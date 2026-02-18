package gitpanel

import (
	"os"
	"path/filepath"
	"strings"
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

func TestMultiLineFieldLinesWrapsAndTruncates(t *testing.T) {
	lines := multiLineFieldLines("commit", "abcdefghijklmnopqrstuvwxyz", 18, 2)
	assert.Len(t, lines, 2)
	assert.True(t, strings.HasPrefix(lines[0], " commit: "))
	assert.True(t, strings.HasPrefix(lines[1], "         "))
	assert.Contains(t, lines[1], "...")
}

func TestGitTabRenderWrapsCommitAndStatus(t *testing.T) {
	m := New()
	m.SetDimensions(42, 10)
	m.SetData(
		"/tmp/repo",
		"main",
		"very-long-commit-subject-that-should-wrap-to-second-line",
		"2026-02-18",
		"donald",
		"modified: file1, file2, file3, file4",
	)
	out := ansi.Strip(m.Render(true))
	assert.Contains(t, out, " commit: very-long-commit-subject-th")
	assert.Contains(t, out, "          -should-wrap-to-second-line")
	assert.Contains(t, out, " status: modified: file1, file2, file")
	assert.Contains(t, out, ", file4")
}

func TestGitTabRenderWrapsStatusToThreeLines(t *testing.T) {
	m := New()
	m.SetDimensions(38, 12)
	m.SetData(
		"/tmp/repo",
		"main",
		"short",
		"2026-02-18",
		"donald",
		"status with many elements to force a third line in panel",
	)
	out := ansi.Strip(m.Render(true))
	assert.Contains(t, out, " status: status with many elements")
	assert.Contains(t, out, "to force a third line")
	assert.Contains(t, out, "panel")
}
