package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	databaseURL := os.Getenv("POSTGRES_URL")
	redisURL := os.Getenv("REDIS_URL")
	searchURL := strings.TrimRight(os.Getenv("ELASTICSEARCH_URL"), "/")
	if databaseURL == "" || redisURL == "" || searchURL == "" {
		return errors.New("POSTGRES_URL, REDIS_URL and ELASTICSEARCH_URL are required")
	}

	db, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return errors.New("invalid PostgreSQL configuration")
	}
	defer db.Close()

	redisOptions, err := redis.ParseURL(redisURL)
	if err != nil {
		return errors.New("invalid Redis configuration")
	}
	cache := redis.NewClient(redisOptions)
	defer cache.Close()

	emailCodes, err := newEmailCodes(cache)
	if err != nil {
		return err
	}
	app := &accountAPI{db: db, cache: cache, emailCodes: emailCodes, secureCookie: os.Getenv("COOKIE_SECURE") != "false"}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/register", app.register)
	mux.HandleFunc("POST /api/auth/register/verify", app.verifyRegistration)
	mux.HandleFunc("POST /api/auth/login", app.login)
	mux.HandleFunc("POST /api/auth/login/verify", app.verifyLogin)
	mux.HandleFunc("POST /api/auth/2fa", app.startEmailTwoFactorChange)
	mux.HandleFunc("POST /api/auth/2fa/verify", app.verifyEmailTwoFactorChange)
	mux.HandleFunc("DELETE /api/auth/session", app.logout)
	mux.HandleFunc("GET /api/me", app.profile)
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL+"/_cluster/health", nil)
		searchReady := false
		if err == nil {
			res, err := http.DefaultClient.Do(req)
			if err == nil {
				res.Body.Close()
				searchReady = res.StatusCode == http.StatusOK
			}
		}
		var usersTableExists bool
		dbErr := db.QueryRow(ctx, `SELECT COUNT(*) = 2 FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = 'users'
			AND column_name IN ('email_verified_at', 'email_2fa_enabled')`).Scan(&usersTableExists)
		storage := map[string]bool{
			"postgres":      dbErr == nil && usersTableExists,
			"redis":         cache.Ping(ctx).Err() == nil,
			"elasticsearch": searchReady,
		}
		status := http.StatusOK
		for _, ready := range storage {
			if !ready {
				status = http.StatusServiceUnavailable
			}
		}
		writeJSON(w, status, storage)
	})

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("Go API listening on %s", server.Addr)
	return server.ListenAndServe()
}
