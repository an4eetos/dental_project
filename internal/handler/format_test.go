package handler

import (
	"strings"
	"testing"

	"github.com/anuarkuanysh/dental_project/internal/model"
)

func TestFormatTelegramReply_analysisOnlyNoDisclaimer(t *testing.T) {
	out := FormatTelegramReply(&model.Analysis{
		VisibleIssues:   []string{"возможное потемнение эмали"},
		Confidence:      "low",
		Recommendations: []string{"чистить зубы два раза в день"},
	})
	if strings.Contains(out, "не медицинская") {
		t.Fatalf("disclaimer must not repeat in every reply: %q", out)
	}
	if !strings.Contains(out, "без диагноза") {
		t.Fatalf("expected non-diagnostic wording: %q", out)
	}
	if !strings.Contains(out, "возможное потемнение эмали") {
		t.Fatalf("expected visible issues in reply: %q", out)
	}
	if !strings.Contains(out, "чистить зубы два раза в день") {
		t.Fatalf("expected recommendations in reply: %q", out)
	}
}
