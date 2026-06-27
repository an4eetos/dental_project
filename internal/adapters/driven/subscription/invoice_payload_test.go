package subscription

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"
)

func TestInvoicePayloadSigner_CompactPayloadFitsTelegramLimit(t *testing.T) {
	t.Parallel()

	signer := NewInvoicePayloadSigner("test-secret-at-least-32-chars-long")
	for _, userID := range []int64{1, 42, 9999999999} {
		payload, err := signer.Sign(userID)
		if err != nil {
			t.Fatalf("Sign(%d): %v", userID, err)
		}
		if len(payload) > invoicePayloadMaxLen {
			t.Fatalf("payload for user %d is %d bytes, limit is %d", userID, len(payload), invoicePayloadMaxLen)
		}
	}
}

func TestInvoicePayloadSigner_TamperedPayloadRejected(t *testing.T) {
	t.Parallel()

	signer := NewInvoicePayloadSigner("test-secret-at-least-32-chars-long")
	payload, err := signer.Sign(5)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	tampered := payload[:len(payload)-1] + "X"
	if _, err := signer.Verify(tampered); err == nil {
		t.Fatal("expected tampered payload to fail")
	}
}

func TestInvoicePayloadSigner_LegacyJSONPayload(t *testing.T) {
	t.Parallel()

	secret := []byte("test-secret-at-least-32-chars-long")
	nonce := make([]byte, 16)
	_, _ = rand.Read(nonce)

	raw, _ := json.Marshal(invoicePayload{
		UserID:   99,
		IssuedAt: time.Now().Unix(),
		Nonce:    base64.RawURLEncoding.EncodeToString(nonce),
	})
	mac := hmac.New(sha256.New, secret)
	mac.Write(raw)
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	encoded := base64.RawURLEncoding.EncodeToString(raw) + "." + sig

	signer := NewInvoicePayloadSigner(string(secret))
	userID, err := signer.Verify(encoded)
	if err != nil {
		t.Fatalf("Verify legacy: %v", err)
	}
	if userID != 99 {
		t.Fatalf("userID = %d", userID)
	}
}

func TestOldJSONPayloadExceedsTelegramLimit(t *testing.T) {
	t.Parallel()

	nonce := make([]byte, 16)
	_, _ = rand.Read(nonce)
	raw, _ := json.Marshal(invoicePayload{
		UserID:   123456789,
		IssuedAt: time.Now().Unix(),
		Nonce:    base64.RawURLEncoding.EncodeToString(nonce),
	})
	mac := hmac.New(sha256.New, []byte("secret"))
	mac.Write(raw)
	encoded := base64.RawURLEncoding.EncodeToString(raw) + "." +
		base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if len(encoded) <= invoicePayloadMaxLen {
		t.Fatalf("expected old payload to exceed limit, got %d bytes", len(encoded))
	}
}
