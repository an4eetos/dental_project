package subscription

import (
	"context"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

// AnswerPreCheckout validates invoice payload and approves Telegram pre-checkout.
type AnswerPreCheckout struct {
	Users    port.UserRepository
	Payments port.TelegramPayments
	Signer   port.InvoicePayloadSigner
}

type PreCheckoutInput struct {
	QueryID string
	Payload string
}

func (uc *AnswerPreCheckout) Execute(ctx context.Context, input PreCheckoutInput) error {
	userID, err := uc.Signer.Verify(input.Payload)
	if err != nil {
		_ = uc.Payments.AnswerPreCheckoutQuery(ctx, input.QueryID, false, "Недействительный платёж")
		return err
	}

	if _, err := uc.Users.GetByID(ctx, userID); err != nil {
		_ = uc.Payments.AnswerPreCheckoutQuery(ctx, input.QueryID, false, "Пользователь не найден")
		return domainerrors.ErrUserNotFound
	}

	return uc.Payments.AnswerPreCheckoutQuery(ctx, input.QueryID, true, "")
}
