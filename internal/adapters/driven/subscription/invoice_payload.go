package subscription

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"strings"
	"time"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

const (
	invoicePayloadMaxAge = 15 * time.Minute
	invoicePayloadMaxLen = 128
	payloadVersion       = byte(1)
	payloadBodyLen       = 1 + 8 + 4 + 8 // version + userID + issuedAt + nonce
	payloadTotalLen      = payloadBodyLen + 8
)

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
	if userID <= 0 {
		return "", domainerrors.ErrInvalidPaymentPayload
	}

	nonce := make([]byte, 8)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	body := make([]byte, payloadBodyLen)
	body[0] = payloadVersion
	binary.BigEndian.PutUint64(body[1:], uint64(userID))
	binary.BigEndian.PutUint32(body[9:], uint32(time.Now().Unix()))
	copy(body[13:], nonce)

	mac := hmac.New(sha256.New, s.secret)
	mac.Write(body)
	token := append(body, mac.Sum(nil)[:8]...)

	encoded := base64.RawURLEncoding.EncodeToString(token)
	if len(encoded) > invoicePayloadMaxLen {
		return "", domainerrors.ErrInvalidPaymentPayload
	}
	return encoded, nil
}

func (s *InvoicePayloadSigner) Verify(encoded string) (int64, error) {
	if userID, err := s.verifyCompact(encoded); err == nil {
		return userID, nil
	}
	return s.verifyLegacy(encoded)
}

func (s *InvoicePayloadSigner) verifyCompact(encoded string) (int64, error) {
	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil || len(raw) != payloadTotalLen || raw[0] != payloadVersion {
		return 0, domainerrors.ErrInvalidPaymentPayload
	}

	body := raw[:payloadBodyLen]
	sig := raw[payloadBodyLen:]

	mac := hmac.New(sha256.New, s.secret)
	mac.Write(body)
	if !hmac.Equal(sig, mac.Sum(nil)[:8]) {
		return 0, domainerrors.ErrInvalidPaymentPayload
	}

	userID := int64(binary.BigEndian.Uint64(body[1:9]))
	if userID <= 0 {
		return 0, domainerrors.ErrInvalidPaymentPayload
	}

	issuedAt := time.Unix(int64(binary.BigEndian.Uint32(body[9:13])), 0)
	if time.Since(issuedAt) > invoicePayloadMaxAge {
		return 0, domainerrors.ErrInvalidPaymentPayload
	}
	return userID, nil
}

func (s *InvoicePayloadSigner) verifyLegacy(encoded string) (int64, error) {
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
