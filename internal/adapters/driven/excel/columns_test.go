package excel

import "testing"

func TestIsIgnorableColumn(t *testing.T) {
	t.Parallel()

	for _, header := range []string{"", "№", "#", "номер", "Unnamed: 0", "1", "  "} {
		if !isIgnorableColumn(header) {
			t.Fatalf("expected ignorable column %q", header)
		}
	}

	if isIgnorableColumn("Ваш возраст (полных лет)?") {
		t.Fatal("expected real column to be kept")
	}
}
