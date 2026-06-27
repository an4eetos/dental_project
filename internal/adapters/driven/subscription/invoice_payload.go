package subscription

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

const invoicePayloadMaxAge = 15 * time.Minute

type invoicePayload struct {
	UserID   int64  `json:"user_id"`
	IssuedAt int64  `json:"issued_at"`
	Nonce    string `json:"nonce"`
}

// InvoicePayloadSigner implements port.InvoicePayloadSigner.
type InvoicePayloadSigner struct {
	secret []byte
}

func NewInvoicePayloadSigner(secret string) *InvoicePayloadSigner {
	return &InvoicePayloadSigner{secret: []byte(secret)}
}

var _ port.InvoicePayloadSigner = (*InvoicePayloadSigner)(nil)

func (s *InvoicePayloadSigner) Sign(userID int64) (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	payload := invoicePayload{
		UserID:   userID,
		IssuedAt: time.Now().Unix(),
		Nonce:    base64.RawURLEncoding.EncodeToString(nonce),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, s.secret)
	mac.Write(raw)
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return base64.RawURLEncoding.EncodeToString(raw) + "." + sig, nil
}

func (s *InvoicePayloadSigner) Verify(encoded string) (int64, error) {
	parts := strings.Split(encoded, ".")
	if len(parts) != 2 {
		return 0, domainerrors.ErrInvalidPaymentPayload
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return 0, domainerrors.ErrInvalidPaymentPayload
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, domainerrors.ErrInvalidPaymentPayload
	}
	mac := hmac.New(sha256.New, s.secret)
	mac.Write(raw)
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return 0, domainerrors.ErrInvalidPaymentPayload
	}
	var payload invoicePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return 0, domainerrors.ErrInvalidPaymentPayload
	}
	if payload.UserID <= 0 {
		return 0, domainerrors.ErrInvalidPaymentPayload
	}
	issued := time.Unix(payload.IssuedAt, 0)
	if time.Since(issued) > invoicePayloadMaxAge {
		return 0, domainerrors.ErrInvalidPaymentPayload
	}
	return payload.UserID, nil
}
