package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

const (
	sessionCookie = "aidm_session"
	sessionTTL    = 12 * time.Hour
)

type accountAPI struct {
	db           *pgxpool.Pool
	cache        *redis.Client
	emailCodes   *emailCodes
	secureCookie bool
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

type accountProfile struct {
	ID             string  `json:"id"`
	Email          string  `json:"email"`
	Nickname       string  `json:"nickname"`
	AvatarURL      *string `json:"avatar_url"`
	EmailVerified  bool    `json:"email_verified"`
	EmailTwoFactor bool    `json:"email_2fa_enabled"`
}

func (a *accountAPI) register(w http.ResponseWriter, r *http.Request) {
	input, ok := readCredentials(w, r)
	if !ok {
		return
	}
	nickname := strings.TrimSpace(input.Nickname)
	if length := utf8.RuneCountInString(nickname); length < 2 || length > 40 {
		writeError(w, http.StatusBadRequest, "Имя должно содержать от 2 до 40 символов.")
		return
	}
	if a.emailCodes == nil {
		writeError(w, http.StatusServiceUnavailable, "Регистрация временно недоступна: не настроена отправка писем.")
		return
	}
	var exists bool
	if err := a.db.QueryRow(r.Context(), "SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)", input.Email).Scan(&exists); err != nil {
		writeError(w, http.StatusServiceUnavailable, "Не удалось начать регистрацию.")
		return
	}
	if exists {
		writeError(w, http.StatusConflict, "Аккаунт с этим адресом уже существует.")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Не удалось создать аккаунт.")
		return
	}
	a.emailCodes.request(w, r, registrationCode, "", emailChallenge{
		Email: input.Email, PasswordHash: string(hash), Nickname: nickname,
	})
}

func (a *accountAPI) verifyRegistration(w http.ResponseWriter, r *http.Request) {
	challenge, ok := a.emailCodes.verify(w, r, registrationCode, "")
	if !ok {
		return
	}
	_, err := a.db.Exec(r.Context(), `
		INSERT INTO users (id, email, password_hash, nickname, email_verified_at)
		VALUES ($1::uuid, $2, $3, $4, NOW())
	`, uuid.NewString(), challenge.Email, challenge.PasswordHash, challenge.Nickname)
	var pgError *pgconn.PgError
	switch {
	case errors.As(err, &pgError) && pgError.Code == "23505" && pgError.ConstraintName == "users_email_key":
		writeError(w, http.StatusConflict, "Аккаунт с этим адресом уже существует.")
	case err != nil:
		writeError(w, http.StatusServiceUnavailable, "Не удалось создать аккаунт.")
	default:
		w.WriteHeader(http.StatusCreated)
	}
}

func (a *accountAPI) login(w http.ResponseWriter, r *http.Request) {
	input, ok := readCredentials(w, r)
	if !ok {
		return
	}
	var userID, hash string
	var emailTwoFactor bool
	err := a.db.QueryRow(r.Context(), `
		SELECT id::text, password_hash, email_2fa_enabled FROM users WHERE email = $1
	`, input.Email).Scan(&userID, &hash, &emailTwoFactor)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusServiceUnavailable, "Не удалось выполнить вход.")
		return
	}
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(input.Password)) != nil {
		writeError(w, http.StatusUnauthorized, "Неверная почта или пароль.")
		return
	}
	if emailTwoFactor {
		a.emailCodes.request(w, r, loginCode, "", emailChallenge{Email: input.Email, UserID: userID})
		return
	}
	a.issueSession(w, r, userID)
}

func (a *accountAPI) verifyLogin(w http.ResponseWriter, r *http.Request) {
	challenge, ok := a.emailCodes.verify(w, r, loginCode, "")
	if !ok {
		return
	}
	var eligible bool
	err := a.db.QueryRow(r.Context(), `
		SELECT EXISTS (
			SELECT 1 FROM users WHERE id = $1::uuid AND email = $2 AND email_2fa_enabled
		)
	`, challenge.UserID, challenge.Email).Scan(&eligible)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "Не удалось выполнить вход.")
		return
	}
	if !eligible {
		writeError(w, http.StatusUnauthorized, "Параметры входа изменены. Войдите заново.")
		return
	}
	a.issueSession(w, r, challenge.UserID)
}

func (a *accountAPI) issueSession(w http.ResponseWriter, r *http.Request, userID string) {
	token := rand.Text()
	if err := a.cache.Set(r.Context(), sessionKey(token), userID, sessionTTL).Err(); err != nil {
		writeError(w, http.StatusServiceUnavailable, "Не удалось создать сессию.")
		return
	}
	if old, err := r.Cookie(sessionCookie); err == nil {
		if err := a.cache.Del(r.Context(), sessionKey(old.Value)).Err(); err != nil {
			_ = a.cache.Del(r.Context(), sessionKey(token)).Err()
			writeError(w, http.StatusServiceUnavailable, "Не удалось заменить сессию.")
			return
		}
	}
	a.setCookie(w, token, int(sessionTTL.Seconds()))
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusNoContent)
}

func (a *accountAPI) authenticatedAccount(w http.ResponseWriter, r *http.Request) (userID, binding string, ok bool) {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Требуется вход.")
		return "", "", false
	}
	binding = sessionKey(cookie.Value)
	userID, err = a.cache.Get(r.Context(), binding).Result()
	if errors.Is(err, redis.Nil) {
		a.setCookie(w, "", -1)
		writeError(w, http.StatusUnauthorized, "Сессия истекла.")
		return "", "", false
	}
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "Не удалось проверить сессию.")
		return "", "", false
	}
	return userID, binding, true
}

func (a *accountAPI) profile(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := a.authenticatedAccount(w, r)
	if !ok {
		return
	}
	var profile accountProfile
	err := a.db.QueryRow(r.Context(), `
		SELECT id::text, email, nickname, avatar_url, email_verified_at IS NOT NULL, email_2fa_enabled
		FROM users WHERE id = $1::uuid
	`, userID).Scan(&profile.ID, &profile.Email, &profile.Nickname, &profile.AvatarURL,
		&profile.EmailVerified, &profile.EmailTwoFactor)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		a.setCookie(w, "", -1)
		writeError(w, http.StatusUnauthorized, "Требуется вход.")
	case err != nil:
		writeError(w, http.StatusServiceUnavailable, "Не удалось загрузить профиль.")
	default:
		writeJSON(w, http.StatusOK, profile)
	}
}

func (a *accountAPI) startEmailTwoFactorChange(w http.ResponseWriter, r *http.Request) {
	userID, binding, ok := a.authenticatedAccount(w, r)
	if !ok {
		return
	}
	var input struct {
		Enabled  *bool  `json:"enabled"`
		Password string `json:"password"`
	}
	if !readJSON(w, r, &input) {
		return
	}
	if input.Enabled == nil || !validPassword(input.Password) {
		writeError(w, http.StatusBadRequest, "Укажите действие и пароль аккаунта.")
		return
	}
	var email, hash string
	var enabled bool
	err := a.db.QueryRow(r.Context(), `
		SELECT email, password_hash, email_2fa_enabled FROM users WHERE id = $1::uuid
	`, userID).Scan(&email, &hash, &enabled)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusUnauthorized, "Требуется вход.")
		return
	}
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "Не удалось проверить настройки аккаунта.")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(input.Password)) != nil {
		writeError(w, http.StatusForbidden, "Неверный пароль.")
		return
	}
	if enabled == *input.Enabled {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	a.emailCodes.request(w, r, settingsCode, binding, emailChallenge{
		Email: email, UserID: userID, Enable: *input.Enabled,
	})
}

func (a *accountAPI) verifyEmailTwoFactorChange(w http.ResponseWriter, r *http.Request) {
	userID, binding, ok := a.authenticatedAccount(w, r)
	if !ok {
		return
	}
	challenge, ok := a.emailCodes.verify(w, r, settingsCode, binding)
	if !ok {
		return
	}
	result, err := a.db.Exec(r.Context(), `
		UPDATE users SET email_2fa_enabled = $3, email_verified_at = COALESCE(email_verified_at, NOW())
		WHERE id = $1::uuid AND email = $2 AND email_2fa_enabled <> $3
	`, userID, challenge.Email, challenge.Enable)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "Не удалось изменить защиту аккаунта.")
		return
	}
	if result.RowsAffected() != 1 {
		writeError(w, http.StatusConflict, "Настройки аккаунта уже изменились. Обновите профиль.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *accountAPI) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookie); err == nil {
		if err := a.cache.Del(r.Context(), sessionKey(cookie.Value)).Err(); err != nil {
			writeError(w, http.StatusServiceUnavailable, "Не удалось завершить сессию.")
			return
		}
	}
	a.setCookie(w, "", -1)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusNoContent)
}

func (a *accountAPI) setCookie(w http.ResponseWriter, token string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookie, Value: token, Path: "/", MaxAge: maxAge,
		Secure: a.secureCookie, HttpOnly: true, SameSite: http.SameSiteLaxMode,
	})
}

func sessionKey(token string) string {
	hash := sha256.Sum256([]byte(token))
	return "aidm:auth:" + hex.EncodeToString(hash[:])
}

func readCredentials(w http.ResponseWriter, r *http.Request) (credentials, bool) {
	var input credentials
	if !readJSON(w, r, &input) {
		return input, false
	}
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	address, err := mail.ParseAddress(input.Email)
	if err != nil || address.Address != input.Email || len(input.Email) > 254 {
		writeError(w, http.StatusBadRequest, "Проверьте адрес электронной почты.")
		return input, false
	}
	if !validPassword(input.Password) {
		writeError(w, http.StatusBadRequest, "Пароль должен содержать от 8 символов и занимать до 72 байт.")
		return input, false
	}
	return input, true
}

func validPassword(password string) bool {
	return utf8.RuneCountInString(password) >= 8 && len(password) <= 72
}

func readJSON(w http.ResponseWriter, r *http.Request, value any) bool {
	mediaType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if mediaType != "application/json" {
		writeError(w, http.StatusUnsupportedMediaType, "Ожидается JSON.")
		return false
	}
	body := http.MaxBytesReader(w, r.Body, 8192)
	defer body.Close()
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil || decoder.Decode(new(any)) != io.EOF {
		writeError(w, http.StatusBadRequest, "Некорректные данные формы.")
		return false
	}
	return true
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
