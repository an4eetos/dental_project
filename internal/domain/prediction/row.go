package prediction

// Row holds keyed string values for inputs or outputs.
type Row map[string]string

// NewInputRow creates an input row with all keys initialized to empty strings.
func NewInputRow() Row {
	return newRow(InputKeys())
}

// NewOutputRow creates an output row with all keys initialized to empty strings.
func NewOutputRow() Row {
	return newRow(OutputKeys())
}

func newRow(keys []string) Row {
	row := make(Row, len(keys))
	for _, key := range keys {
		row[key] = ""
	}
	return row
}
