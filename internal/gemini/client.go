package gemini

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/anuarkuanysh/dental_project/internal/model"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

const geminiGenerateURL = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent"

// Client calls the Gemini REST API for multimodal generation (informational prompts only).
type Client struct {
	httpClient *http.Client
	apiKey     string
	model      string
	log        *slog.Logger
}

var _ port.VisionAnalyzer = (*Client)(nil)

// New constructs a Gemini REST client.
func New(httpClient *http.Client, apiKey, model string, log *slog.Logger) *Client {
	if log == nil {
		log = slog.Default()
	}
	return &Client{httpClient: httpClient, apiKey: apiKey, model: model, log: log}
}

type generateRequest struct {
	Contents         []contentPayload `json:"contents"`
	GenerationConfig genConfig      `json:"generationConfig"`
}

type contentPayload struct {
	Role  string `json:"role,omitempty"`
	Parts []part `json:"parts"`
}

type part struct {
	Text       string       `json:"text,omitempty"`
	InlineData *inlineData  `json:"inline_data,omitempty"`
}

type inlineData struct {
	MIMEType string `json:"mime_type"`
	Data     string `json:"data"`
}

type genConfig struct {
	Temperature      float32 `json:"temperature"`
	ResponseMIMEType string  `json:"responseMimeType"`
}

type generateResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

// AnalyzeTeethImage sends an image to Gemini and expects JSON matching model.Analysis.
func (c *Client) AnalyzeTeethImage(ctx context.Context, image []byte, mimeType string) (*model.Analysis, error) {
	if len(image) == 0 {
		return nil, errors.New("empty image")
	}

	b64 := base64.StdEncoding.EncodeToString(image)

	body := generateRequest{
		Contents: []contentPayload{
			{
				Role: "user",
				Parts: []part{
					{
						InlineData: &inlineData{
							MIMEType: mimeType,
							Data:     b64,
						},
					},
					{Text: teethVisionPrompt()},
				},
			},
		},
		GenerationConfig: genConfig{
			Temperature:      0.35,
			ResponseMIMEType: "application/json",
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	base := fmt.Sprintf(geminiGenerateURL, c.model)
	u, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("parse gemini url: %w", err)
	}
	q := u.Query()
	q.Set("key", c.apiKey)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		c.log.Warn("gemini non-OK response", "status", resp.StatusCode, "body", truncate(string(raw), 800))
		return nil, fmt.Errorf("gemini status %d", resp.StatusCode)
	}

	var gr generateResponse
	if err := json.Unmarshal(raw, &gr); err != nil {
		return nil, fmt.Errorf("decode gemini envelope: %w", err)
	}

	if gr.Error != nil {
		return nil, fmt.Errorf("gemini api error: %s (%s)", gr.Error.Message, gr.Error.Status)
	}

	if len(gr.Candidates) == 0 || len(gr.Candidates[0].Content.Parts) == 0 {
		return nil, errors.New("gemini returned no candidates")
	}

	text := strings.TrimSpace(gr.Candidates[0].Content.Parts[0].Text)
	text = extractJSON(text)

	var out model.Analysis
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		c.log.Warn("failed to parse analysis json", "text", truncate(text, 1200), "err", err)
		return nil, fmt.Errorf("parse analysis json: %w", err)
	}

	normalizeAnalysis(&out)
	return &out, nil
}

func teethVisionPrompt() string {
	return `You analyze a single dental photo for general, non-diagnostic educational purposes.

STRICT RULES:
- Do NOT diagnose diseases or conditions. Avoid medical certainty.
- Describe only potentially visible issues using cautious Russian wording (e.g. «возможно», «может выглядеть как»).
- Acknowledge uncertainty and limits of one photo, lighting, and camera quality.
- Hygiene suggestions only — no treatment plans.
- Informational only; not medical advice.
- The traffic-light level is a rough, non-clinical hint for the user — never state certainty.

TRAFFIC LIGHT (informational triage only — assign exactly one):
- "green": photo quality allows a rough look and nothing clearly worrying is visible; gums/teeth may appear generally okay. Routine hygiene is enough for now. Still not a clean bill of health.
- "yellow": possible mild concerns worth monitoring — e.g. plaque buildup, slight gum redness, mild staining, uneven color; schedule a routine dental check when convenient.
- "red": signs that reasonably suggest prompt professional evaluation — e.g. obvious swelling, heavy bleeding appearance, severe discoloration, broken tooth, pus, large cavity-like dark area, or strong concern despite photo limits. Urge dentist visit soon; not an emergency diagnosis.

If the image is too blurry, dark, or not showing teeth/gums clearly, prefer "yellow", lower confidence, and say so in traffic_light_summary.

LANGUAGE: All strings except traffic_light and confidence tokens MUST be in Russian.

Respond ONLY with valid JSON (no markdown fences) exactly in this shape:
{
  "traffic_light": "green|yellow|red",
  "traffic_light_summary": "одно-два предложения на русском, осторожно, без диагноза",
  "visible_issues": ["строка на русском"],
  "confidence": "low|medium|high",
  "recommendations": ["строка на русском"],
  "disclaimer": "Это не медицинская консультация."
}

Use English tokens only for "traffic_light" (green|yellow|red) and "confidence" (low|medium|high).
Keep lists concise (max ~6 items each).`
}

func normalizeAnalysis(a *model.Analysis) {
	if a == nil {
		return
	}
	switch strings.ToLower(strings.TrimSpace(a.TrafficLight)) {
	case "green", "yellow", "red":
		a.TrafficLight = strings.ToLower(strings.TrimSpace(a.TrafficLight))
	default:
		a.TrafficLight = "yellow"
	}
	if strings.TrimSpace(a.TrafficLightSummary) == "" {
		a.TrafficLightSummary = defaultTrafficSummary(a.TrafficLight)
	}
	if strings.TrimSpace(a.Disclaimer) == "" {
		a.Disclaimer = "Это не медицинская консультация."
	}
}

func defaultTrafficSummary(level string) string {
	switch level {
	case "green":
		return "По снимку явных тревожных признаков не видно; поддерживайте гигиену и плановые осмотры."
	case "red":
		return "Есть признаки, из‑за которых стоит обратиться к стоматологу в ближайшее время."
	default:
		return "Есть нюансы, которые лучше обсудить со стоматологом на плановом приёме."
	}
}

func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimPrefix(s, "json")
		s = strings.TrimSpace(s)
		if idx := strings.LastIndex(s, "```"); idx != -1 {
			s = strings.TrimSpace(s[:idx])
		}
	}
	return s
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
