package imageproc

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"

	"github.com/disintegration/imaging"

	"github.com/anuarkuanysh/dental_project/internal/port"
)

// Processor resizes large images to reduce payload size for Gemini.
type Processor struct {
	MaxDimension int
}

var _ port.ImageProcessor = (*Processor)(nil)

// New creates an image processor with the given maximum edge length in pixels.
func New(maxDimension int) *Processor {
	if maxDimension <= 0 {
		maxDimension = 1024
	}
	return &Processor{MaxDimension: maxDimension}
}

// PrepareForVision decodes the image, optionally downsamples, and returns JPEG bytes.
func (p *Processor) PrepareForVision(imageBytes []byte, _ string) ([]byte, string, error) {
	img, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return nil, "", fmt.Errorf("decode image: %w", err)
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= 0 || h <= 0 {
		return nil, "", fmt.Errorf("invalid image dimensions")
	}

	out := img
	longest := max(w, h)
	if longest > p.MaxDimension {
		out = imaging.Fit(img, p.MaxDimension, p.MaxDimension, imaging.Lanczos)
	}

	var buf bytes.Buffer
	if err := imaging.Encode(&buf, out, imaging.JPEG, imaging.JPEGQuality(85)); err != nil {
		return nil, "", fmt.Errorf("encode jpeg: %w", err)
	}

	return buf.Bytes(), "image/jpeg", nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
