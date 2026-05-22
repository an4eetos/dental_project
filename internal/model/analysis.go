package model

// Analysis is the structured JSON returned by Gemini (informational only).
type Analysis struct {
	// TrafficLight is an informational triage hint: green | yellow | red (not a diagnosis).
	TrafficLight        string   `json:"traffic_light"`
	TrafficLightSummary string   `json:"traffic_light_summary"`
	VisibleIssues       []string `json:"visible_issues"`
	Confidence       string   `json:"confidence"`
	Recommendations  []string `json:"recommendations"`
	Disclaimer       string   `json:"disclaimer"`
}
