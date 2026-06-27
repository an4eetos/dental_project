package subscription

import (
	"context"
	"time"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
	"github.com/anuarkuanysh/dental_project/internal/port"
	subscriptionservice "github.com/anuarkuanysh/dental_project/internal/service/subscription"
)

// PlanConfig holds subscription pricing exposed to clients.
type PlanConfig struct {
	StarsPrice   int
	Duration     time.Duration
	InvoiceTitle string
	InvoiceDesc  string
}

// CreateInvoice creates a Telegram Stars invoice link for the authenticated user.
type CreateInvoice struct {
	Users    port.UserRepository
	Payments port.TelegramPayments
	Signer   port.InvoicePayloadSigner
	Plan     PlanConfig
}

type CreateInvoiceResult struct {
	InvoiceLink string
}

func (uc *CreateInvoice) Execute(ctx context.Context, userID int64) (CreateInvoiceResult, error) {
	user, err := uc.Users.GetByID(ctx, userID)
	if err != nil {
		return CreateInvoiceResult{}, err
	}
	if user.Role != identity.RolePatient {
		return CreateInvoiceResult{}, domainerrors.ErrForbidden
	}
	if uc.Plan.StarsPrice <= 0 {
		return CreateInvoiceResult{}, domainerrors.ErrInvoiceLinkFailed
	}

	payload, err := uc.Signer.Sign(userID)
	if err != nil {
		return CreateInvoiceResult{}, err
	}
	if len(payload) > subscriptionservice.MaxInvoicePayloadLen {
		return CreateInvoiceResult{}, domainerrors.ErrInvalidPaymentPayload
	}

	link, err := uc.Payments.CreateInvoiceLink(ctx, port.CreateInvoiceLinkParams{
		Title:       subscriptionservice.TruncateInvoiceTitle(uc.Plan.InvoiceTitle),
		Description: subscriptionservice.TruncateInvoiceDescription(uc.Plan.InvoiceDesc),
		Payload:     payload,
		Currency:    "XTR",
		Prices: []port.LabeledPrice{{
			Label:  subscriptionservice.TruncateInvoiceTitle(uc.Plan.InvoiceTitle),
			Amount: uc.Plan.StarsPrice,
		}},
	})
	if err != nil {
		return CreateInvoiceResult{}, err
	}
	return CreateInvoiceResult{InvoiceLink: link}, nil
}
