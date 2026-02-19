package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatMetadataLines_ModifiedOnTwoLines(t *testing.T) {
	meta := [][2]string{
		{keyDataModified, "2026-02-19 21:12:33"},
	}

	lines := formatMetadataLines(meta, 0, 4, len(keyDataModified), 19)

	assert.Len(t, lines, 2)
	assert.Equal(t, "Modified 2026-02-19", lines[0])
	assert.Equal(t, "         21:12:33", lines[1])
}
