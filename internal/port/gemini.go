package port

import (
	"context"

	"github.com/anuarkuanysh/dental_project/internal/model"
)

// VisionAnalyzer analyzes dental-related photos using Gemini (informational output only).
type VisionAnalyzer interface {
	AnalyzeTeethImage(ctx context.Context, image []byte, mimeType string) (*model.Analysis, error)
}
