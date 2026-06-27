package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/anuarkuanysh/dental_project/internal/port"
)

// PaymentsClient implements port.TelegramPayments via raw Bot API HTTP calls.
type PaymentsClient struct {
	token  string
	httpCl *http.Client
}

func NewPaymentsClient(token string, httpClient *http.Client) *PaymentsClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &PaymentsClient{token: token, httpCl: httpClient}
}

var _ port.TelegramPayments = (*PaymentsClient)(nil)

type labeledPrice struct {
	Label  string `json:"label"`
	Amount int    `json:"amount"`
}

type createInvoiceLinkRequest struct {
	Title         string         `json:"title"`
	Description   string         `json:"description"`
	Payload       string         `json:"payload"`
	ProviderToken string         `json:"provider_token"`
	Currency      string         `json:"currency"`
	Prices        []labeledPrice `json:"prices"`
}

type answerPreCheckoutRequest struct {
	PreCheckoutQueryID string `json:"pre_checkout_query_id"`
	OK                 bool   `json:"ok"`
	ErrorMessage       string `json:"error_message,omitempty"`
}

type apiResponse struct {
	OK          bool            `json:"ok"`
	Result      json.RawMessage `json:"result"`
	Description string          `json:"description"`
}

func (c *PaymentsClient) CreateInvoiceLink(ctx context.Context, params port.CreateInvoiceLinkParams) (string, error) {
	prices := make([]labeledPrice, len(params.Prices))
	for i, p := range params.Prices {
		prices[i] = labeledPrice{Label: p.Label, Amount: p.Amount}
	}
	body := createInvoiceLinkRequest{
		Title:         params.Title,
		Description:   params.Description,
		Payload:       params.Payload,
		ProviderToken: "",
		Currency:      params.Currency,
		Prices:        prices,
	}
	var link string
	if err := c.call(ctx, "createInvoiceLink", body, &link); err != nil {
		return "", err
	}
	return link, nil
}

func (c *PaymentsClient) AnswerPreCheckoutQuery(ctx context.Context, queryID string, ok bool, errorMessage string) error {
	body := answerPreCheckoutRequest{
		PreCheckoutQueryID: queryID,
		OK:                 ok,
		ErrorMessage:       errorMessage,
	}
	return c.call(ctx, "answerPreCheckoutQuery", body, nil)
}

func (c *PaymentsClient) call(ctx context.Context, method string, payload any, result any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", c.token, method)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpCl.Do(req)
	if err != nil {
		return fmt.Errorf("telegram api: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	var apiResp apiResponse
	if err := json.Unmarshal(raw, &apiResp); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if !apiResp.OK {
		if apiResp.Description != "" {
			return fmt.Errorf("telegram api %s: %s", method, apiResp.Description)
		}
		return fmt.Errorf("telegram api %s failed", method)
	}
	if result != nil && len(apiResp.Result) > 0 {
		if err := json.Unmarshal(apiResp.Result, result); err != nil {
			return fmt.Errorf("decode result: %w", err)
		}
	}
	return nil
}
