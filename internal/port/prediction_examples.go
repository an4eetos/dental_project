package port

import (
	"context"

	"github.com/anuarkuanysh/dental_project/internal/domain/prediction"
)

// PredictionExampleRepository provides labeled examples for few-shot prompting.
// Excel and database adapters can implement this interface interchangeably.
type PredictionExampleRepository interface {
	ListExamples(ctx context.Context) ([]prediction.Example, error)
}
