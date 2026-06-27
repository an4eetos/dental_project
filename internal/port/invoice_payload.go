package port

// InvoicePayloadSigner signs and verifies Telegram invoice payloads.
type InvoicePayloadSigner interface {
	Sign(userID int64) (string, error)
	Verify(payload string) (int64, error)
}
