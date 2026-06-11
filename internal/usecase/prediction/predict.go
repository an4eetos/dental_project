package prediction

import (
	"context"
	"strings"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	domain "github.com/anuarkuanysh/dental_project/internal/domain/prediction"
	promptsvc "github.com/anuarkuanysh/dental_project/internal/service/prediction"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

// Predict generates outputs from survey inputs using few-shot Gemini prompting.
type Predict struct {
	Examples  port.PredictionExampleRepository
	Generator port.TextGenerator
}

// Execute loads examples, builds a few-shot prompt, and returns the generated outputs.
func (uc *Predict) Execute(ctx context.Context, input domain.Row) (domain.Row, error) {
	examples, err := uc.Examples.ListExamples(ctx)
	if err != nil {
		return nil, err
	}
	if len(examples) == 0 {
		return nil, domainerrors.ErrNoPredictionExamples
	}

	normalized := domain.NewInputRow()
	for _, key := range domain.InputKeys() {
		normalized[key] = strings.TrimSpace(input[key])
	}

	prompt := promptsvc.BuildFewShotPrompt(examples, normalized)
	raw, err := uc.Generator.GenerateJSON(ctx, prompt)
	if err != nil {
		return nil, domainerrors.ErrPredictionFailed
	}

	output, err := promptsvc.ParseOutputJSON(raw)
	if err != nil {
		return nil, domainerrors.ErrPredictionFailed
	}

	return output, nil
}
