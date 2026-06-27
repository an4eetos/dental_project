package port

import "context"

// LabeledPrice is a single line item on a Telegram invoice.
type LabeledPrice struct {
	Label  string
	Amount int
}

// CreateInvoiceLinkParams builds a Stars invoice link for the Mini App.
type CreateInvoiceLinkParams struct {
	Title       string
	Description string
	Payload     string
	Currency    string
	Prices      []LabeledPrice
}

// TelegramPayments creates invoice links and answers pre-checkout queries.
type TelegramPayments interface {
	CreateInvoiceLink(ctx context.Context, params CreateInvoiceLinkParams) (string, error)
	AnswerPreCheckoutQuery(ctx context.Context, queryID string, ok bool, errorMessage string) error
}
