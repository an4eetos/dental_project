package gemini

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
)

func (c *Client) newGenerateRequest(ctx context.Context, payload []byte) (*http.Request, error) {
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
	return req, nil
}
