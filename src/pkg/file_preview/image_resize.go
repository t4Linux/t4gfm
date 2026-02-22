package filepreview

import (
	"bytes"
	"fmt"
	"image"
	"log/slog"

	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
)

// prepareImageForPreview handles the complete image preparation pipeline
func prepareImageForPreview(data []byte) (image.Image, int, int, error) {
	imgReader := bytes.NewReader(data)

	img, _, err := image.Decode(imgReader)
	if err != nil {
		return nil, 0, 0, err
	}

	// Store original dimensions
	originalWidth := img.Bounds().Dx()
	originalHeight := img.Bounds().Dy()

	// Adjust orientation based on EXIF data
	exifReader := bytes.NewReader(data)
	img = adjustImageOrientation(exifReader, img)

	// Limit resolution to 1080p
	img = limitImageResolution(img, originalWidth, originalHeight)

	return img, originalWidth, originalHeight, nil
}

// limitImageResolution limits image resolution to 1080p while maintaining aspect ratio
func limitImageResolution(img image.Image, originalWidth, originalHeight int) image.Image {
	const maxImageWidth = 1920
	const maxImageHeight = 1080

	// Only resize if the image is larger than 1080p
	if originalWidth > maxImageWidth || originalHeight > maxImageHeight {
		resizedImg := imaging.Fit(img, maxImageWidth, maxImageHeight, imaging.Lanczos)
		return resizedImg
	}

	return img
}

// adjustImageOrientation adjusts image orientation based on EXIF data
func adjustImageOrientation(r *bytes.Reader, img image.Image) image.Image {
	exifData, err := exif.Decode(r)
	if err != nil {
		slog.Error("exif error", "error", err)
		return img
	}
	tag, err := exifData.Get(exif.Orientation)
	if err != nil {
		slog.Error("exif orientation error", "error", err)
		return img
	}
	orientation, err := tag.Int(0)
	if err != nil {
		slog.Error("exif orientation value error", "error", err)
		return img
	}
	return adjustOrientation(img, orientation)
}

// adjustOrientation applies the specified orientation transformation to the image
func adjustOrientation(img image.Image, orientation int) image.Image {
	switch orientation {
	case 1:
		return img
	case 2: //nolint:mnd // EXIF orientation: horizontal flip
		return imaging.FlipH(img)
	case 3: //nolint:mnd // EXIF orientation: 180 rotation
		return imaging.Rotate180(img)
	case 4: //nolint:mnd // EXIF orientation: vertical flip
		return imaging.FlipV(img)
	case 5: //nolint:mnd // EXIF orientation: transpose
		return imaging.Transpose(img)
	case 6: //nolint:mnd // EXIF orientation: 270 rotation
		return imaging.Rotate270(img)
	case 7: //nolint:mnd // EXIF orientation: transverse
		return imaging.Transverse(img)
	case 8: //nolint:mnd // EXIF orientation: 90 rotation
		return imaging.Rotate90(img)
	default:
		slog.Error("Invalid orientation value", "error", orientation)
		return img
	}
}

// resizeForANSI resizes image specifically for ANSI rendering
func resizeForANSI(img image.Image, maxWidth, maxHeight, cellWidth, cellHeight int) image.Image {
	width, height := computeANSITargetDimensions(
		img.Bounds().Dx(),
		img.Bounds().Dy(),
		maxWidth,
		maxHeight,
		cellWidth,
		cellHeight,
	)
	if width <= 0 || height <= 0 {
		slog.Error("Invalid ANSI resize dimensions",
			"width", width,
			"height", height,
			"maxWidth", maxWidth,
			"maxHeight", maxHeight,
			"cellWidth", cellWidth,
			"cellHeight", cellHeight)
		// Use maxHeight*2 because each terminal row represents 2 pixel rows in ANSI rendering
		return imaging.Fit(img, maxWidth, maxHeight*heightScaleFactor, imaging.Lanczos)
	}
	return imaging.Resize(img, width, height, imaging.Lanczos)
}

func computeANSITargetDimensions(
	originalWidth, originalHeight int,
	maxWidth, maxHeight int,
	cellWidth, cellHeight int,
) (int, int) {
	if originalWidth <= 0 || originalHeight <= 0 || maxWidth <= 0 || maxHeight <= 0 {
		return 0, 0
	}
	if cellWidth <= 0 {
		cellWidth = DefaultPixelsPerColumn
	}
	if cellHeight <= 0 {
		cellHeight = DefaultPixelsPerRow
	}

	maxANSIHeight := maxHeight * heightScaleFactor

	displayAdjustedRatio := (float64(originalWidth) / float64(originalHeight)) *
		(float64(cellHeight) / (float64(heightScaleFactor) * float64(cellWidth)))

	boxRatio := float64(maxWidth) / float64(maxANSIHeight)

	var targetWidth, targetHeight int
	if displayAdjustedRatio > boxRatio {
		targetWidth = maxWidth
		targetHeight = int(float64(targetWidth) / displayAdjustedRatio)
	} else {
		targetHeight = maxANSIHeight
		targetWidth = int(float64(targetHeight) * displayAdjustedRatio)
	}

	if targetWidth < 1 {
		targetWidth = 1
	}
	if targetHeight < 1 {
		targetHeight = 1
	}

	if targetWidth > maxWidth {
		return maxWidth, maxANSIHeight
	}
	if targetHeight > maxANSIHeight {
		return maxWidth, maxANSIHeight
	}

	if targetWidth <= 0 || targetHeight <= 0 {
		slog.Error("computeANSITargetDimensions generated invalid result",
			"originalWidth", originalWidth,
			"originalHeight", originalHeight,
			"maxWidth", maxWidth,
			"maxHeight", maxHeight,
			"cellWidth", cellWidth,
			"cellHeight", cellHeight,
			"target", fmt.Sprintf("%dx%d", targetWidth, targetHeight))
		return maxWidth, maxANSIHeight
	}

	return targetWidth, targetHeight
}
