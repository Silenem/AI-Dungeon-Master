package main

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/redis/go-redis/v9"
)

const emailCodeTTL = 10 * time.Minute

type emailPurpose string

const (
	registrationCode emailPurpose = "registration"
	loginCode        emailPurpose = "login"
	settingsCode     emailPurpose = "settings"
)

func (p emailPurpose) caption() string {
	switch p {
	case registrationCode:
		return "подтверждение регистрации"
	case loginCode:
		return "подтверждение входа"
	default:
		return "изменение защиты аккаунта"
	}
}

type emailChallenge struct {
	Email        string `json:"email,omitempty"`
	PasswordHash string `json:"password_hash,omitempty"`
	Nickname     string `json:"nickname,omitempty"`
	UserID       string `json:"user_id,omitempty"`
	Enable       bool   `json:"enable,omitempty"`
}

type emailCodes struct {
	cache  *redis.Client
	sender *smtpSender
	secret []byte
}

func newEmailCodes(cache *redis.Client) (*emailCodes, error) {
	sender, err := loadSMTPSender()
	if err != nil || sender == nil {
		return nil, err
	}
	secret, err := hex.DecodeString(os.Getenv("AUTH_CODE_SECRET"))
	if err != nil || len(secret) != 32 {
		return nil, errors.New("AUTH_CODE_SECRET must contain 32 bytes in hexadecimal")
	}
	return &emailCodes{cache: cache, sender: sender, secret: secret}, nil
}

var emailSendQuota = redis.NewScript(`
if redis.call('EXISTS', KEYS[1]) == 1 or
   tonumber(redis.call('GET', KEYS[2]) or '0') >= 8 or
   tonumber(redis.call('GET', KEYS[3]) or '0') >= 90 then
	return 0
end
redis.call('SET', KEYS[1], 1, 'EX', 45)
if redis.call('INCR', KEYS[2]) == 1 then redis.call('EXPIRE', KEYS[2], 3600) end
if redis.call('INCR', KEYS[3]) == 1 then redis.call('EXPIRE', KEYS[3], 600) end
return 1
`)

func (e *emailCodes) request(w http.ResponseWriter, r *http.Request, purpose emailPurpose, binding string, challenge emailChallenge) {
	if e == nil {
		writeError(w, http.StatusServiceUnavailable, "Отправка писем пока недоступна.")
		return
	}
	ctx := r.Context()
	recipient := e.digest("recipient", challenge.Email)
	keys := []string{
		"aidm:email:cooldown:" + string(purpose) + ":" + recipient,
		"aidm:email:recipient:" + recipient,
		"aidm:email:global",
	}
	reserved, err := emailSendQuota.Run(ctx, e.cache, keys).Int()
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "Не удалось создать код.")
		return
	}
	if reserved != 1 {
		writeError(w, http.StatusTooManyRequests, "Слишком много запросов кода. Попробуйте позже.")
		return
	}
	number, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "Не удалось создать код.")
		return
	}
	id := rand.Text()
	code := fmt.Sprintf("%06d", number.Int64())
	payload, err := json.Marshal(challenge)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "Не удалось создать код.")
		return
	}
	key := "aidm:email:challenge:" + id
	_, err = e.cache.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.HSet(ctx, key, "purpose", string(purpose), "binding", binding,
			"digest", e.digest(id, code), "payload", payload)
		pipe.Expire(ctx, key, emailCodeTTL)
		return nil
	})
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "Не удалось создать код.")
		return
	}
	if err := e.sender.sendCode(ctx, challenge.Email, code, purpose); err != nil {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second)
		defer cancel()
		_ = e.cache.Del(cleanup, key).Err()
		writeError(w, http.StatusServiceUnavailable, "Не удалось отправить письмо с кодом.")
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"challenge_id": id})
}

var consumeEmailCode = redis.NewScript(`
local fields = redis.call('HMGET', KEYS[1], 'purpose', 'binding', 'digest', 'payload')
if fields[1] ~= ARGV[1] or fields[2] ~= ARGV[2] then return false end
if fields[3] ~= ARGV[3] then
	if redis.call('HINCRBY', KEYS[1], 'failures', 1) >= 5 then redis.call('DEL', KEYS[1]) end
	return false
end
redis.call('DEL', KEYS[1])
return fields[4]
`)

var (
	emailChallengeID = regexp.MustCompile(`^[A-Z2-7]{26,128}$`)
	emailCode        = regexp.MustCompile(`^[0-9]{6}$`)
)

func (e *emailCodes) verify(w http.ResponseWriter, r *http.Request, purpose emailPurpose, binding string) (emailChallenge, bool) {
	var challenge emailChallenge
	if e == nil {
		writeError(w, http.StatusServiceUnavailable, "Подтверждение почты пока недоступно.")
		return challenge, false
	}
	var input struct {
		ChallengeID string `json:"challenge_id"`
		Code        string `json:"code"`
	}
	if !readJSON(w, r, &input) {
		return challenge, false
	}
	if !emailChallengeID.MatchString(input.ChallengeID) || !emailCode.MatchString(input.Code) {
		writeError(w, http.StatusBadRequest, "Введите шестизначный код из письма.")
		return challenge, false
	}
	payload, err := consumeEmailCode.Run(r.Context(), e.cache,
		[]string{"aidm:email:challenge:" + input.ChallengeID},
		string(purpose), binding, e.digest(input.ChallengeID, input.Code)).Text()
	if errors.Is(err, redis.Nil) {
		writeError(w, http.StatusBadRequest, "Код неверен или истёк. При необходимости запросите новый.")
		return challenge, false
	}
	if err != nil || json.Unmarshal([]byte(payload), &challenge) != nil {
		writeError(w, http.StatusServiceUnavailable, "Не удалось проверить код.")
		return challenge, false
	}
	return challenge, true
}

func (e *emailCodes) digest(label, value string) string {
	mac := hmac.New(sha256.New, e.secret)
	mac.Write([]byte(label + "\x00" + value))
	return hex.EncodeToString(mac.Sum(nil))
}
