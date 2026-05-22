package handler

import (
	"strings"
	"testing"

	"github.com/anuarkuanysh/dental_project/internal/model"
)

func TestFormatTelegramReply_containsDisclaimer(t *testing.T) {
	out := FormatTelegramReply(&model.Analysis{
		TrafficLight:        "yellow",
		TrafficLightSummary: "Возможен налёт; стоит проверить у врача.",
		VisibleIssues:       []string{"возможное потемнение эмали"},
		Confidence:          "low",
		Recommendations:     []string{"чистить зубы два раза в день"},
		Disclaimer:          "Это не медицинская консультация.",
	})
	if !strings.Contains(out, "не медицинская") {
		t.Fatalf("expected disclaimer language: %q", out)
	}
	if !strings.Contains(out, "без диагноза") {
		t.Fatalf("expected non-diagnostic wording: %q", out)
	}
	if !strings.Contains(out, "🟡") {
		t.Fatalf("expected traffic light in reply: %q", out)
	}
}
