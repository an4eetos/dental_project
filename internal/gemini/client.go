package gemini

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
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
var _ port.TextGenerator = (*Client)(nil)

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

// AnalyzeTeethImage sends photo or video bytes to Gemini and expects JSON matching model.Analysis.
func (c *Client) AnalyzeTeethImage(ctx context.Context, image []byte, mimeType string) (*model.Analysis, error) {
	if len(image) == 0 {
		return nil, errors.New("empty media")
	}

	b64 := base64.StdEncoding.EncodeToString(image)
	prompt := teethVisionPrompt()
	if strings.HasPrefix(strings.ToLower(mimeType), "video/") {
		prompt = teethVideoPrompt()
	}

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
					{Text: prompt},
				},
			},
		},
		GenerationConfig: genConfig{
			Temperature:      0.35,
			ResponseMIMEType: "application/json",
		},
	}

	text, err := c.generateContent(ctx, body)
	if err != nil {
		return nil, err
	}
	text = extractJSON(text)

	var out model.Analysis
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		c.log.Warn("failed to parse analysis json", "text", truncate(text, 1200), "err", err)
		return nil, fmt.Errorf("parse analysis json: %w", err)
	}

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

If the image is too blurry, dark, or not showing teeth/gums clearly, lower confidence and mention it in visible_issues.

LANGUAGE: All strings except confidence tokens MUST be in Russian.

Respond ONLY with valid JSON (no markdown fences) exactly in this shape:
{
  "visible_issues": ["строка на русском"],
  "confidence": "low|medium|high",
  "recommendations": ["строка на русском"]
}

Use English tokens only for "confidence" (low|medium|high).
Keep lists concise (max ~6 items each).`
}

func teethVideoPrompt() string {
	return `You analyze a short dental video for general, non-diagnostic educational purposes.

STRICT RULES:
- Do NOT diagnose diseases or conditions. Avoid medical certainty.
- Describe only potentially visible issues using cautious Russian wording (e.g. «возможно», «может выглядеть как»).
- Acknowledge uncertainty and limits of one video, lighting, movement, and camera quality.
- Hygiene suggestions only — no treatment plans.
- Informational only; not medical advice.

If the video is too blurry, dark, or not showing teeth/gums clearly, lower confidence and mention it in visible_issues.

LANGUAGE: All strings except confidence tokens MUST be in Russian.

Respond ONLY with valid JSON (no markdown fences) exactly in this shape:
{
  "visible_issues": ["строка на русском"],
  "confidence": "low|medium|high",
  "recommendations": ["строка на русском"]
}

Use English tokens only for "confidence" (low|medium|high).
Keep lists concise (max ~6 items each).`
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
