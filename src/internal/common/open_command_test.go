package common

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldPreferSystemOpen(t *testing.T) {
	assert.True(t, ShouldPreferSystemOpen("/tmp/file.pdf"))
	assert.True(t, ShouldPreferSystemOpen("/tmp/image.JPG"))
	assert.True(t, ShouldPreferSystemOpen("/tmp/movie.mkv"))
	assert.False(t, ShouldPreferSystemOpen("/tmp/code.go"))
	assert.False(t, ShouldPreferSystemOpen("/tmp/notes.txt"))
}

func TestResolveSystemOpenCommand(t *testing.T) {
	cmd := ResolveSystemOpenCommand()
	switch runtime.GOOS {
	case "darwin":
		assert.Equal(t, "open", cmd)
	case "windows":
		assert.Equal(t, "start", cmd)
	default:
		assert.Equal(t, "xdg-open", cmd)
	}
}

func TestResolveEnterOpenProgram(t *testing.T) {
	oldOpenWith := Config.OpenWith
	oldEditor := Config.Editor
	t.Cleanup(func() {
		Config.OpenWith = oldOpenWith
		Config.Editor = oldEditor
	})

	Config.Editor = "nvim"
	Config.OpenWith = map[string]string{"pdf": "zathura"}

	assert.Equal(t, "zathura", ResolveEnterOpenProgram("/tmp/file.pdf"))
	assert.Equal(t, ResolveSystemOpenCommand(), ResolveEnterOpenProgram("/tmp/image.jpg"))
	assert.Equal(t, "nvim", ResolveEnterOpenProgram("/tmp/main.go"))
}
