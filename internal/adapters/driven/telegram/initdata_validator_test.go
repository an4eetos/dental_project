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
	"testing"
	"time"
)

func TestInitDataValidator_Validate(t *testing.T) {
	t.Parallel()

	botToken := "123456:ABC-DEF"
	user := map[string]any{
		"id":         int64(42),
		"first_name": "Иван",
		"username":   "ivan",
	}
	userJSON, _ := json.Marshal(user)
	authDate := time.Now().Unix()

	values := url.Values{}
	values.Set("auth_date", strconv.FormatInt(authDate, 10))
	values.Set("user", string(userJSON))
	values.Set("query_id", "AAE")

	hash := signInitData(values, botToken)
	values.Set("hash", hash)

	v := NewInitDataValidator(botToken, time.Hour)
	profile, err := v.Validate(context.Background(), values.Encode())
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if profile.TelegramID != 42 || profile.FirstName != "Иван" {
		t.Fatalf("unexpected profile: %+v", profile)
	}
}

func signInitData(values url.Values, botToken string) string {
	var pairs []string
	for key, vals := range values {
		if key == "hash" || len(vals) == 0 {
			continue
		}
		pairs = append(pairs, key+"="+vals[0])
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	secretMAC := hmac.New(sha256.New, []byte("WebAppData"))
	secretMAC.Write([]byte(botToken))
	secretKey := secretMAC.Sum(nil)

	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(dataCheckString))
	return hex.EncodeToString(mac.Sum(nil))
}
