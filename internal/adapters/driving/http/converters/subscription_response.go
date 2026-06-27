package converters

import (
	"time"

	"github.com/anuarkuanysh/dental_project/internal/domain/subscription"
)

type SubscriptionStatusResponse struct {
	Active       bool    `json:"active"`
	ExpiresAt    *string `json:"expires_at,omitempty"`
	StarsPrice   int     `json:"stars_price"`
	DurationDays int     `json:"duration_days"`
}

type CreateInvoiceResponse struct {
	InvoiceLink string `json:"invoice_link"`
}

func ToSubscriptionStatusResponse(entitlement subscription.Entitlement) SubscriptionStatusResponse {
	resp := SubscriptionStatusResponse{
		Active:       entitlement.Active,
		StarsPrice:   entitlement.StarsPrice,
		DurationDays: entitlement.DurationDays,
	}
	if entitlement.ExpiresAt != nil {
		formatted := entitlement.ExpiresAt.UTC().Format(time.RFC3339)
		resp.ExpiresAt = &formatted
	}
	return resp
}
