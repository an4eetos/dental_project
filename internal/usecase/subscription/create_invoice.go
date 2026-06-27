package subscription

import (
	"context"
	"time"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
	"github.com/anuarkuanysh/dental_project/internal/port"
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

	payload, err := uc.Signer.Sign(userID)
	if err != nil {
		return CreateInvoiceResult{}, err
	}

	link, err := uc.Payments.CreateInvoiceLink(ctx, port.CreateInvoiceLinkParams{
		Title:       uc.Plan.InvoiceTitle,
		Description: uc.Plan.InvoiceDesc,
		Payload:     payload,
		Currency:    "XTR",
		Prices: []port.LabeledPrice{{
			Label:  uc.Plan.InvoiceTitle,
			Amount: uc.Plan.StarsPrice,
		}},
	})
	if err != nil {
		return CreateInvoiceResult{}, err
	}
	return CreateInvoiceResult{InvoiceLink: link}, nil
}
