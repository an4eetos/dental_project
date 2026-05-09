package model

// Analysis is the structured JSON returned by Gemini (informational only).
type Analysis struct {
	VisibleIssues    []string `json:"visible_issues"`
	Confidence       string   `json:"confidence"`
	Recommendations  []string `json:"recommendations"`
	Disclaimer       string   `json:"disclaimer"`
}
