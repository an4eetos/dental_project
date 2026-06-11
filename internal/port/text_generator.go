package port

import "context"

// TextGenerator produces completions from a prompt.
type TextGenerator interface {
	GenerateText(ctx context.Context, prompt string) (string, error)
	GenerateJSON(ctx context.Context, prompt string) (string, error)
}
