package filepreview

import (
	"errors"
	"fmt"
	"image/color"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Terminal cell to pixel conversion constants
// These approximate the pixel dimensions of terminal cells
const (
	DefaultPixelsPerColumn = 10 // approximate pixels per terminal column
	DefaultPixelsPerRow    = 20 // approximate pixels per terminal row
)

// TerminalCellSize represents the pixel dimensions of terminal cells
type TerminalCellSize struct {
	PixelsPerColumn int
	PixelsPerRow    int
}

// TerminalCapabilities encapsulates terminal capability detection
type TerminalCapabilities struct {
	cellSize       TerminalCellSize
	cellSizeInit   sync.Once
	detectionMutex sync.Mutex
}

// NewTerminalCapabilities creates a new TerminalCapabilities instance
func NewTerminalCapabilities() *TerminalCapabilities {
	return &TerminalCapabilities{
		cellSize: TerminalCellSize{
			PixelsPerColumn: DefaultPixelsPerColumn,
			PixelsPerRow:    DefaultPixelsPerRow,
		},
	}
}

// InitTerminalCapabilities initializes all terminal capabilities detection
// including cell size and Kitty Graphics Protocol support
// This should be called early in the application startup
func (tc *TerminalCapabilities) InitTerminalCapabilities() {
	// Use a goroutine to avoid blocking the application startup
	go func() {
		// Initialize cell size detection
		tc.cellSizeInit.Do(func() {
			tc.cellSize = tc.detectTerminalCellSize()
			slog.Info("Terminal cell size detection",
				"pixels_per_column", tc.cellSize.PixelsPerColumn,
				"pixels_per_row", tc.cellSize.PixelsPerRow)
		})
	}()
}

// GetTerminalCellSize returns the current terminal cell size
// If detection hasn't been initialized, it performs detection first
func (tc *TerminalCapabilities) GetTerminalCellSize() TerminalCellSize {
	tc.cellSizeInit.Do(func() {
		tc.cellSize = tc.detectTerminalCellSize()
		slog.Info("Terminal cell size detection (lazy init)",
			"pixels_per_column", tc.cellSize.PixelsPerColumn,
			"pixels_per_row", tc.cellSize.PixelsPerRow)
	})

	return tc.cellSize
}

// winsize struct for ioctl TIOCGWINSZ
type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

// detectTerminalCellSize detects the terminal cell size using ioctl system calls
// This method is non-blocking and doesn't interfere with stdin
func (tc *TerminalCapabilities) detectTerminalCellSize() TerminalCellSize {
	tc.detectionMutex.Lock()
	defer tc.detectionMutex.Unlock()

	if cellSize, ok := getTerminalCellSizeFromEnv(); ok {
		slog.Info("Using terminal cell size from env override",
			"pixels_per_column", cellSize.PixelsPerColumn,
			"pixels_per_row", cellSize.PixelsPerRow)
		return cellSize
	}

	if cellSize, ok := getTerminalCellSizeViaIoctl(); ok {
		slog.Info("Successfully detected terminal cell size via ioctl",
			"pixels_per_column", cellSize.PixelsPerColumn,
			"pixels_per_row", cellSize.PixelsPerRow)
		return cellSize
	}

	// Fallback to default values
	slog.Info("Using default terminal cell size")
	return getDefaultCellSize()
}

// getDefaultCellSize returns default fallback terminal cell size
func getDefaultCellSize() TerminalCellSize {
	return TerminalCellSize{
		PixelsPerColumn: DefaultPixelsPerColumn,
		PixelsPerRow:    DefaultPixelsPerRow,
	}
}

func getTerminalCellSizeFromEnv() (TerminalCellSize, bool) {
	override := strings.TrimSpace(os.Getenv("T4GFM_PREVIEW_CELL_SIZE"))
	if override != "" {
		if cellSize, err := parseCellSizeOverride(override); err == nil {
			return cellSize, true
		}
	}

	widthValue := strings.TrimSpace(os.Getenv("T4GFM_PREVIEW_CELL_WIDTH"))
	heightValue := strings.TrimSpace(os.Getenv("T4GFM_PREVIEW_CELL_HEIGHT"))
	if widthValue == "" || heightValue == "" {
		return TerminalCellSize{}, false
	}

	width, widthErr := strconv.Atoi(widthValue)
	height, heightErr := strconv.Atoi(heightValue)
	if widthErr != nil || heightErr != nil || width <= 0 || height <= 0 {
		return TerminalCellSize{}, false
	}

	return TerminalCellSize{PixelsPerColumn: width, PixelsPerRow: height}, true
}

func parseCellSizeOverride(value string) (TerminalCellSize, error) {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == 'x' || r == 'X' || r == ':' || r == ',' || r == ';'
	})
	if len(parts) != 2 {
		return TerminalCellSize{}, fmt.Errorf("invalid override format")
	}

	width, widthErr := strconv.Atoi(strings.TrimSpace(parts[0]))
	height, heightErr := strconv.Atoi(strings.TrimSpace(parts[1]))
	if widthErr != nil || heightErr != nil || width <= 0 || height <= 0 {
		return TerminalCellSize{}, fmt.Errorf("invalid override values")
	}

	return TerminalCellSize{PixelsPerColumn: width, PixelsPerRow: height}, nil
}

// InitTerminalCapabilities initializes terminal capabilities for the ImagePreviewer
func (p *ImagePreviewer) InitTerminalCapabilities() {
	p.terminalCap.InitTerminalCapabilities()
}

func hexToColor(hex string) (color.RGBA, error) {
	if len(hex) != 7 || hex[0] != '#' {
		return color.RGBA{}, errors.New("invalid hex color format")
	}
	values, err := strconv.ParseUint(hex[1:], 16, 32)
	if err != nil {
		return color.RGBA{}, err
	}
	return color.RGBA{
		R: uint8(values >> rgbShift16),            //nolint:gosec // RGB values are masked to 8-bit range
		G: uint8((values >> rgbShift8) & rgbMask), //nolint:gosec // RGB values are masked to 8-bit range
		B: uint8(values & rgbMask),                //nolint:gosec // RGB values are masked to 8-bit range
		A: alphaOpaque,
	}, nil
}

func colorToHex(color color.Color) string {
	r, g, b, _ := color.RGBA()
	return fmt.Sprintf(
		"#%02x%02x%02x",
		uint8(r>>rgbShift8), //nolint:gosec // RGBA() returns 16-bit values, shifting by 8 gives 8-bit
		uint8(g>>rgbShift8), //nolint:gosec // RGBA() returns 16-bit values, shifting by 8 gives 8-bit
		uint8(b>>rgbShift8), //nolint:gosec // RGBA() returns 16-bit values, shifting by 8 gives 8-bit
	)
}
