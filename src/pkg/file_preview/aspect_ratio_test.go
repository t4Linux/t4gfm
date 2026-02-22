package filepreview

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeANSITargetDimensionsUsesCellRatio(t *testing.T) {
	widthA, heightA := computeANSITargetDimensions(1080, 1920, 80, 40, 10, 20)
	widthB, heightB := computeANSITargetDimensions(1080, 1920, 80, 40, 8, 20)

	assert.Equal(t, 45, widthA)
	assert.Equal(t, 56, widthB)
	assert.Equal(t, 80, heightA)
	assert.Equal(t, 80, heightB)
}

func TestParseCellSizeOverride(t *testing.T) {
	cellSize, err := parseCellSizeOverride("9x18")
	assert.NoError(t, err)
	assert.Equal(t, 9, cellSize.PixelsPerColumn)
	assert.Equal(t, 18, cellSize.PixelsPerRow)

	cellSize, err = parseCellSizeOverride("10:22")
	assert.NoError(t, err)
	assert.Equal(t, 10, cellSize.PixelsPerColumn)
	assert.Equal(t, 22, cellSize.PixelsPerRow)

	_, err = parseCellSizeOverride("bad")
	assert.Error(t, err)
}

func TestGetTerminalCellSizeFromEnv(t *testing.T) {
	originalCombined, hadCombined := os.LookupEnv("T4GFM_PREVIEW_CELL_SIZE")
	originalWidth, hadWidth := os.LookupEnv("T4GFM_PREVIEW_CELL_WIDTH")
	originalHeight, hadHeight := os.LookupEnv("T4GFM_PREVIEW_CELL_HEIGHT")
	t.Cleanup(func() {
		if hadCombined {
			_ = os.Setenv("T4GFM_PREVIEW_CELL_SIZE", originalCombined)
		} else {
			_ = os.Unsetenv("T4GFM_PREVIEW_CELL_SIZE")
		}
		if hadWidth {
			_ = os.Setenv("T4GFM_PREVIEW_CELL_WIDTH", originalWidth)
		} else {
			_ = os.Unsetenv("T4GFM_PREVIEW_CELL_WIDTH")
		}
		if hadHeight {
			_ = os.Setenv("T4GFM_PREVIEW_CELL_HEIGHT", originalHeight)
		} else {
			_ = os.Unsetenv("T4GFM_PREVIEW_CELL_HEIGHT")
		}
	})

	_ = os.Setenv("T4GFM_PREVIEW_CELL_SIZE", "11x23")
	_ = os.Unsetenv("T4GFM_PREVIEW_CELL_WIDTH")
	_ = os.Unsetenv("T4GFM_PREVIEW_CELL_HEIGHT")

	cellSize, ok := getTerminalCellSizeFromEnv()
	assert.True(t, ok)
	assert.Equal(t, 11, cellSize.PixelsPerColumn)
	assert.Equal(t, 23, cellSize.PixelsPerRow)

	_ = os.Unsetenv("T4GFM_PREVIEW_CELL_SIZE")
	_ = os.Setenv("T4GFM_PREVIEW_CELL_WIDTH", "12")
	_ = os.Setenv("T4GFM_PREVIEW_CELL_HEIGHT", "24")

	cellSize, ok = getTerminalCellSizeFromEnv()
	assert.True(t, ok)
	assert.Equal(t, 12, cellSize.PixelsPerColumn)
	assert.Equal(t, 24, cellSize.PixelsPerRow)
}
