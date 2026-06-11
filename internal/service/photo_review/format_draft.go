package photoreview

import (
	"fmt"
	"strings"

	"github.com/anuarkuanysh/dental_project/internal/model"
)

// FormatAnalysisDraft turns a Gemini analysis into editable Russian text for doctors.
func FormatAnalysisDraft(a *model.Analysis) string {
	if a == nil {
		return ""
	}

	var b strings.Builder
	b.WriteString("Возможные заметные особенности (черновик ИИ, без диагноза):\n")
	if len(a.VisibleIssues) == 0 {
		b.WriteString("- надёжно не выявлено\n")
	} else {
		for _, item := range a.VisibleIssues {
			b.WriteString(fmt.Sprintf("- %s\n", item))
		}
	}

	b.WriteString(fmt.Sprintf("\nУверенность модели / качества снимка: %s\n", confidenceLabelRU(a.Confidence)))

	b.WriteString("\nОбщие советы по гигиене (не план лечения):\n")
	if len(a.Recommendations) == 0 {
		b.WriteString("- поддерживайте регулярную чистку зубов и использование нити; персональные рекомендации — у врача\n")
	} else {
		for _, item := range a.Recommendations {
			b.WriteString(fmt.Sprintf("- %s\n", item))
		}
	}
	return b.String()
}

func confidenceLabelRU(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "low":
		return "низкая"
	case "medium":
		return "средняя"
	case "high":
		return "высокая"
	default:
		return "неизвестно"
	}
}
