package subscription

import (
	"strings"
	"testing"
)

func TestTruncateInvoiceTitle(t *testing.T) {
	t.Parallel()

	title := strings.Repeat("а", 40)
	got := TruncateInvoiceTitle(title)
	if len([]rune(got)) != MaxInvoiceTitleLen {
		t.Fatalf("got %d runes, want %d", len([]rune(got)), MaxInvoiceTitleLen)
	}
}

func TestTruncateInvoiceDescription(t *testing.T) {
	t.Parallel()

	desc := strings.Repeat("b", 300)
	got := TruncateInvoiceDescription(desc)
	if len([]rune(got)) != MaxInvoiceDescriptionLen {
		t.Fatalf("got %d runes, want %d", len([]rune(got)), MaxInvoiceDescriptionLen)
	}
}
