package excel

import (
	"context"

	domain "github.com/anuarkuanysh/dental_project/internal/domain/prediction"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

// EmptyRepository is used when no examples file is available.
type EmptyRepository struct{}

var _ port.PredictionExampleRepository = (*EmptyRepository)(nil)

// NewEmptyRepository returns a repository with no few-shot examples.
func NewEmptyRepository() *EmptyRepository {
	return &EmptyRepository{}
}

// ListExamples returns an empty slice.
func (r *EmptyRepository) ListExamples(_ context.Context) ([]domain.Example, error) {
	return nil, nil
}
