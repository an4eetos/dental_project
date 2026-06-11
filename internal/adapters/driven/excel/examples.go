package excel

import (
	"context"
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"

	domain "github.com/anuarkuanysh/dental_project/internal/domain/prediction"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

// Repository loads few-shot examples from an Excel file at startup.
type Repository struct {
	examples []domain.Example
}

var _ port.PredictionExampleRepository = (*Repository)(nil)

// NewRepository reads and parses examples from the given .xlsx path.
func NewRepository(path string) (*Repository, error) {
	examples, err := loadExamples(path)
	if err != nil {
		return nil, err
	}
	return &Repository{examples: examples}, nil
}

// ListExamples returns the cached examples loaded from Excel.
func (r *Repository) ListExamples(_ context.Context) ([]domain.Example, error) {
	out := make([]domain.Example, len(r.examples))
	copy(out, r.examples)
	return out, nil
}

func loadExamples(path string) ([]domain.Example, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("open excel file %q: %w", path, err)
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	if sheet == "" {
		return nil, fmt.Errorf("excel file %q has no sheets", path)
	}

	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("read rows from %q: %w", path, err)
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("excel file %q must contain a header row and at least one example", path)
	}

	colIndex, err := parseHeader(rows[0])
	if err != nil {
		return nil, err
	}

	examples := make([]domain.Example, 0, len(rows)-1)
	for i, row := range rows[1:] {
		ex, ok, err := parseRow(row, colIndex)
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", i+2, err)
		}
		if ok {
			examples = append(examples, ex)
		}
	}

	if len(examples) == 0 {
		return nil, fmt.Errorf("excel file %q contains no valid example rows", path)
	}

	return examples, nil
}

type columnMapping struct {
	inputKeys  map[int]string
	outputKeys map[int]string
}

func parseHeader(header []string) (columnMapping, error) {
	mapping := columnMapping{
		inputKeys:  make(map[int]string),
		outputKeys: make(map[int]string),
	}

	for i, name := range header {
		key := strings.TrimSpace(name)
		if isIgnorableColumn(key) {
			continue
		}
		if inputKey, ok := domain.ColumnToInputKey(key); ok {
			mapping.inputKeys[i] = inputKey
			continue
		}
		if outputKey, ok := domain.ColumnToOutputKey(key); ok {
			mapping.outputKeys[i] = outputKey
			continue
		}
		return columnMapping{}, fmt.Errorf("unknown column %q", key)
	}

	for _, field := range domain.InputFields {
		found := false
		for _, key := range mapping.inputKeys {
			if key == field.Key {
				found = true
				break
			}
		}
		if !found {
			return columnMapping{}, fmt.Errorf("missing required column %q", field.Column)
		}
	}

	for _, field := range domain.OutputFields {
		found := false
		for _, key := range mapping.outputKeys {
			if key == field.Key {
				found = true
				break
			}
		}
		if !found {
			return columnMapping{}, fmt.Errorf("missing required column %q", field.Column)
		}
	}

	return mapping, nil
}

func parseRow(row []string, colIndex columnMapping) (domain.Example, bool, error) {
	get := func(idx int) string {
		if idx >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[idx])
	}

	input := domain.NewInputRow()
	for idx, key := range colIndex.inputKeys {
		input[key] = get(idx)
	}

	expected := domain.NewOutputRow()
	for idx, key := range colIndex.outputKeys {
		expected[key] = get(idx)
	}

	empty := true
	for _, v := range input {
		if v != "" {
			empty = false
			break
		}
	}
	for _, v := range expected {
		if v != "" {
			empty = false
			break
		}
	}
	if empty {
		return domain.Example{}, false, nil
	}

	for _, field := range domain.OutputFields {
		if expected[field.Key] == "" {
			return domain.Example{}, false, fmt.Errorf("column %q is required", field.Column)
		}
	}

	return domain.Example{Input: input, Expected: expected}, true, nil
}
