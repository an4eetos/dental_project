package port

// ImageProcessor optionally resizes/compresses images before sending to Gemini.
type ImageProcessor interface {
	PrepareForVision(image []byte, mimeHint string) ([]byte, string, error)
}
