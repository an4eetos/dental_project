package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestPaymentsClient_CreateInvoiceLink_OmitsProviderTokenForStars(t *testing.T) {
	t.Parallel()

	client := NewPaymentsClient("TEST", &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if !strings.Contains(req.URL.Path, "createInvoiceLink") {
				t.Fatalf("unexpected path: %s", req.URL.Path)
			}

			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}
			var captured map[string]json.RawMessage
			if err := json.Unmarshal(body, &captured); err != nil {
				t.Fatalf("unmarshal body: %v", err)
			}
			if _, ok := captured["provider_token"]; ok {
				t.Fatalf("provider_token must be omitted for XTR, body=%s", string(body))
			}

			respBody := `{"ok":true,"result":"https://t.me/$testinvoice"}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(respBody)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		}),
	})

	link, err := client.CreateInvoiceLink(context.Background(), port.CreateInvoiceLinkParams{
		Title:       "Подписка",
		Description: "Видео",
		Payload:     "compact-payload",
		Currency:    "XTR",
		Prices:      []port.LabeledPrice{{Label: "Подписка", Amount: 50}},
	})
	if err != nil {
		t.Fatalf("CreateInvoiceLink: %v", err)
	}
	if link != "https://t.me/$testinvoice" {
		t.Fatalf("link = %q", link)
	}
}

func TestPaymentsClient_CreateInvoiceLink_MapsTelegramError(t *testing.T) {
	t.Parallel()

	client := NewPaymentsClient("TEST", &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			respBody := `{"ok":false,"description":"PAYLOAD_INVALID"}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(respBody)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		}),
	})

	_, err := client.CreateInvoiceLink(context.Background(), port.CreateInvoiceLinkParams{
		Title:       "Test",
		Description: "Test",
		Payload:     "x",
		Currency:    "XTR",
		Prices:      []port.LabeledPrice{{Label: "Test", Amount: 1}},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, domainerrors.ErrInvoiceLinkFailed) {
		t.Fatalf("err = %v", err)
	}
}
