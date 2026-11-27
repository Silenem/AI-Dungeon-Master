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
	secureCookie bool
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

type accountProfile struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Nickname  string  `json:"nickname"`
	AvatarURL *string `json:"avatar_url"`
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
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Не удалось создать аккаунт.")
		return
	}
	_, err = a.db.Exec(r.Context(), `
		INSERT INTO users (id, email, password_hash, nickname)
		VALUES ($1::uuid, $2, $3, $4)
	`, uuid.NewString(), input.Email, string(hash), nickname)
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
	err := a.db.QueryRow(r.Context(), `
		SELECT id::text, password_hash FROM users WHERE email = $1
	`, input.Email).Scan(&userID, &hash)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusServiceUnavailable, "Не удалось выполнить вход.")
		return
	}
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(input.Password)) != nil {
		writeError(w, http.StatusUnauthorized, "Неверная почта или пароль.")
		return
	}

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

func (a *accountAPI) profile(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Требуется вход.")
		return
	}
	userID, err := a.cache.Get(r.Context(), sessionKey(cookie.Value)).Result()
	if errors.Is(err, redis.Nil) {
		a.setCookie(w, "", -1)
		writeError(w, http.StatusUnauthorized, "Сессия истекла.")
		return
	}
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "Не удалось проверить сессию.")
		return
	}

	var profile accountProfile
	err = a.db.QueryRow(r.Context(), `
		SELECT id::text, email, nickname, avatar_url FROM users WHERE id = $1::uuid
	`, userID).Scan(&profile.ID, &profile.Email, &profile.Nickname, &profile.AvatarURL)
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
	mediaType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if mediaType != "application/json" {
		writeError(w, http.StatusUnsupportedMediaType, "Ожидается JSON.")
		return input, false
	}
	body := http.MaxBytesReader(w, r.Body, 8192)
	defer body.Close()
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil || decoder.Decode(new(any)) != io.EOF {
		writeError(w, http.StatusBadRequest, "Некорректные данные формы.")
		return input, false
	}
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	address, err := mail.ParseAddress(input.Email)
	if err != nil || address.Address != input.Email || len(input.Email) > 254 {
		writeError(w, http.StatusBadRequest, "Проверьте адрес электронной почты.")
		return input, false
	}
	if utf8.RuneCountInString(input.Password) < 8 || len(input.Password) > 72 {
		writeError(w, http.StatusBadRequest, "Пароль должен содержать от 8 символов и занимать до 72 байт.")
		return input, false
	}
	return input, true
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
