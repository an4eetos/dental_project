package prediction

// Example is a labeled training row used for few-shot prompting.
type Example struct {
	Input    Row
	Expected Row
}
