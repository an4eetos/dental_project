package excel_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"

	domain "github.com/anuarkuanysh/dental_project/internal/domain/prediction"
	exceladapter "github.com/anuarkuanysh/dental_project/internal/adapters/driven/excel"
)

func TestRepository_loadsExamplesWithLeadingIndexColumn(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "examples-index.xlsx")
	writeSampleWorkbookWithIndexColumn(t, path)

	repo, err := exceladapter.NewRepository(path)
	if err != nil {
		t.Fatalf("NewRepository: %v", err)
	}

	examples, err := repo.ListExamples(context.Background())
	if err != nil {
		t.Fatalf("ListExamples: %v", err)
	}
	if len(examples) != 2 {
		t.Fatalf("expected 2 examples, got %d", len(examples))
	}
	if examples[0].Input[domain.KeyAge] != "28" {
		t.Fatalf("unexpected age: %q", examples[0].Input[domain.KeyAge])
	}
}

func TestRepository_loadsExamplesFromExcel(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "examples.xlsx")
	writeSampleWorkbook(t, path)

	repo, err := exceladapter.NewRepository(path)
	if err != nil {
		t.Fatalf("NewRepository: %v", err)
	}

	examples, err := repo.ListExamples(context.Background())
	if err != nil {
		t.Fatalf("ListExamples: %v", err)
	}
	if len(examples) != 2 {
		t.Fatalf("expected 2 examples, got %d", len(examples))
	}
	if examples[0].Expected[domain.KeyRiskGroup] != "средняя" {
		t.Fatalf("unexpected first risk group: %q", examples[0].Expected[domain.KeyRiskGroup])
	}
}

func writeSampleWorkbookWithIndexColumn(t *testing.T, path string) {
	t.Helper()

	headers := append([]string{""}, append(inputHeaders(), outputHeaders()...)...)
	writeWorkbook(t, path, headers, [][]string{
		{"1", "28", "20", "5", "2", "да", "да", "6.8", "45%", "средняя", "наблюдение", "контроль через 3 месяца"},
		{"2", "32", "30", "2", "3", "нет", "нет", "7.1", "20%", "низкая", "профилактика", "фторирование"},
	})
}

func writeSampleWorkbook(t *testing.T, path string) {
	t.Helper()

	headers := append(inputHeaders(), outputHeaders()...)
	writeWorkbook(t, path, headers, [][]string{
		{"28", "20", "5", "2", "да", "да", "6.8", "45%", "средняя", "наблюдение", "контроль через 3 месяца"},
		{"32", "30", "2", "3", "нет", "нет", "7.1", "20%", "низкая", "профилактика", "фторирование"},
	})
}

func writeWorkbook(t *testing.T, path string, headers []string, rows [][]string) {
	t.Helper()

	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheet, cell, h); err != nil {
			t.Fatalf("set header: %v", err)
		}
	}

	for r, row := range rows {
		for c, value := range row {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
			if err := f.SetCellValue(sheet, cell, value); err != nil {
				t.Fatalf("set cell: %v", err)
			}
		}
	}

	if err := f.SaveAs(path); err != nil {
		t.Fatalf("save workbook: %v", err)
	}
}

func inputHeaders() []string {
	headers := make([]string, len(domain.InputFields))
	for i, field := range domain.InputFields {
		headers[i] = field.Column
	}
	return headers
}

func outputHeaders() []string {
	headers := make([]string, len(domain.OutputFields))
	for i, field := range domain.OutputFields {
		headers[i] = field.Column
	}
	return headers
}
