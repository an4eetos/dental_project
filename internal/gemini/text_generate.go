package gemini

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// GenerateText sends a text-only prompt to Gemini and returns the raw completion.
func (c *Client) GenerateText(ctx context.Context, prompt string) (string, error) {
	if strings.TrimSpace(prompt) == "" {
		return "", errors.New("empty prompt")
	}

	body := generateRequest{
		Contents: []contentPayload{
			{
				Role:  "user",
				Parts: []part{{Text: prompt}},
			},
		},
		GenerationConfig: genConfig{
			Temperature: 0.2,
		},
	}

	text, err := c.generateContent(ctx, body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

// GenerateJSON sends a prompt and requests a JSON object response.
func (c *Client) GenerateJSON(ctx context.Context, prompt string) (string, error) {
	if strings.TrimSpace(prompt) == "" {
		return "", errors.New("empty prompt")
	}

	body := generateRequest{
		Contents: []contentPayload{
			{
				Role:  "user",
				Parts: []part{{Text: prompt}},
			},
		},
		GenerationConfig: genConfig{
			Temperature:      0.2,
			ResponseMIMEType: "application/json",
		},
	}

	text, err := c.generateContent(ctx, body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

func (c *Client) generateContent(ctx context.Context, body generateRequest) (string, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := c.newGenerateRequest(ctx, payload)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		c.log.Warn("gemini non-OK response", "status", resp.StatusCode, "body", truncate(string(raw), 800))
		return "", fmt.Errorf("gemini status %d", resp.StatusCode)
	}

	var gr generateResponse
	if err := json.Unmarshal(raw, &gr); err != nil {
		return "", fmt.Errorf("decode gemini envelope: %w", err)
	}

	if gr.Error != nil {
		return "", fmt.Errorf("gemini api error: %s (%s)", gr.Error.Message, gr.Error.Status)
	}

	if len(gr.Candidates) == 0 || len(gr.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("gemini returned no candidates")
	}

	return strings.TrimSpace(gr.Candidates[0].Content.Parts[0].Text), nil
}
