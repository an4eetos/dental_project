package prediction

import (
	"encoding/json"
	"fmt"
	"strings"

	domain "github.com/anuarkuanysh/dental_project/internal/domain/prediction"
)

const systemInstructions = `You are a dental pregnancy risk assistant. Given patient survey answers, predict four clinical outputs by following the patterns in the examples.

Rules:
- Study the examples carefully and infer the mapping from inputs to outputs.
- Return ONLY a JSON object with exactly these keys: %s
- No explanation, markdown, or extra keys.
- Match the style and format of the example output values.`

// BuildFewShotPrompt formats labeled examples and the target input into a JSON-completion prompt.
func BuildFewShotPrompt(examples []domain.Example, input domain.Row) string {
	keysJSON, _ := json.Marshal(domain.OutputKeys())

	var b strings.Builder
	fmt.Fprintf(&b, systemInstructions, string(keysJSON))
	b.WriteString("\n\n")

	for i, ex := range examples {
		fmt.Fprintf(&b, "Example %d:\n", i+1)
		writeInputBlock(&b, ex.Input)
		writeOutputBlock(&b, ex.Expected)
		b.WriteString("\n")
	}

	b.WriteString("Now predict for:\n")
	writeInputBlock(&b, input)
	b.WriteString("JSON:")

	return b.String()
}

func writeInputBlock(b *strings.Builder, row domain.Row) {
	for _, field := range domain.InputFields {
		fmt.Fprintf(b, "%s: %s\n", field.Column, strings.TrimSpace(row[field.Key]))
	}
}

func writeOutputBlock(b *strings.Builder, row domain.Row) {
	for _, field := range domain.OutputFields {
		fmt.Fprintf(b, "%s: %s\n", field.Column, strings.TrimSpace(row[field.Key]))
	}
}
