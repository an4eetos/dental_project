package telegram

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
)

type telegramUserJSON struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	PhotoURL  string `json:"photo_url"`
}

// InitDataValidator validates Telegram Web App initData signatures.
type InitDataValidator struct {
	BotToken string
	MaxAge   time.Duration
}

func NewInitDataValidator(botToken string, maxAge time.Duration) *InitDataValidator {
	return &InitDataValidator{BotToken: botToken, MaxAge: maxAge}
}

func (v *InitDataValidator) Validate(_ context.Context, initData string) (identity.TelegramProfile, error) {
	if strings.TrimSpace(initData) == "" {
		return identity.TelegramProfile{}, domainerrors.ErrInvalidInitData
	}

	values, err := url.ParseQuery(initData)
	if err != nil {
		return identity.TelegramProfile{}, domainerrors.ErrInvalidInitData
	}

	receivedHash := values.Get("hash")
	if receivedHash == "" {
		return identity.TelegramProfile{}, domainerrors.ErrInvalidInitData
	}
	values.Del("hash")

	if !v.verifyHash(values, receivedHash) {
		return identity.TelegramProfile{}, domainerrors.ErrInvalidInitData
	}

	authDateRaw := values.Get("auth_date")
	if authDateRaw == "" {
		return identity.TelegramProfile{}, domainerrors.ErrInvalidInitData
	}
	authUnix, err := strconv.ParseInt(authDateRaw, 10, 64)
	if err != nil {
		return identity.TelegramProfile{}, domainerrors.ErrInvalidInitData
	}
	authTime := time.Unix(authUnix, 0).UTC()
	if time.Since(authTime) > v.MaxAge {
		return identity.TelegramProfile{}, domainerrors.ErrInvalidInitData
	}

	userRaw := values.Get("user")
	if userRaw == "" {
		return identity.TelegramProfile{}, domainerrors.ErrInvalidInitData
	}

	var tu telegramUserJSON
	if err := json.Unmarshal([]byte(userRaw), &tu); err != nil {
		return identity.TelegramProfile{}, domainerrors.ErrInvalidInitData
	}
	if tu.ID == 0 || strings.TrimSpace(tu.FirstName) == "" {
		return identity.TelegramProfile{}, domainerrors.ErrInvalidInitData
	}

	return identity.TelegramProfile{
		TelegramID: tu.ID,
		Username:   tu.Username,
		FirstName:  tu.FirstName,
		LastName:   tu.LastName,
		AvatarURL:  tu.PhotoURL,
	}, nil
}

func (v *InitDataValidator) verifyHash(values url.Values, receivedHash string) bool {
	var pairs []string
	for key, vals := range values {
		if len(vals) == 0 {
			continue
		}
		pairs = append(pairs, key+"="+vals[0])
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	secretMAC := hmac.New(sha256.New, []byte("WebAppData"))
	secretMAC.Write([]byte(v.BotToken))
	secretKey := secretMAC.Sum(nil)

	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(dataCheckString))
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(receivedHash))
}
