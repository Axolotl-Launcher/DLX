package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/OwO-Network/DLX/sponsor-gateway/internal/afdian"
	"github.com/OwO-Network/DLX/sponsor-gateway/internal/api"
	"github.com/OwO-Network/DLX/sponsor-gateway/internal/auth"
	"github.com/OwO-Network/DLX/sponsor-gateway/internal/store"
	"github.com/OwO-Network/DLX/sponsor-gateway/internal/translate"
	"github.com/OwO-Network/DLX/sponsor-gateway/internal/usage"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
)

func secretFromEnv(name string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	path := os.Getenv(name + "_FILE")
	if path == "" {
		return ""
	}
	value, err := os.ReadFile(path)
	if err != nil {
		log.Printf("unable to read configured secret file for %s", name)
		return ""
	}
	return strings.TrimSpace(string(value))
}
func requiredEnv(name string) string {
	value := secretFromEnv(name)
	if value == "" {
		log.Fatalf("%s must be configured", name)
	}
	return value
}

func main() {
	address := os.Getenv("GATEWAY_ADDR")
	if address == "" {
		address = ":8080"
	}
	dlxURL := os.Getenv("DLX_INTERNAL_URL")
	if dlxURL == "" {
		dlxURL = "http://dlx:1188"
	}
	hasher, err := auth.NewHasher(requiredEnv("API_KEY_PEPPER"))
	if err != nil {
		log.Fatal("API_KEY_PEPPER must be configured with at least 32 bytes")
	}
	sessions, err := auth.NewSessions(requiredEnv("SESSION_SECRET"), 7*24*time.Hour)
	if err != nil {
		log.Fatal("SESSION_SECRET must be configured with at least 32 bytes")
	}
	db, err := sql.Open("pgx", requiredEnv("DATABASE_URL"))
	if err != nil {
		log.Fatal("unable to create PostgreSQL connection")
	}
	defer db.Close()
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		log.Fatal("PostgreSQL is unavailable")
	}
	if err = store.Migrate(ctx, db); err != nil {
		log.Fatal("database migration failed")
	}
	redisClient := redis.NewClient(&redis.Options{Addr: requiredEnv("REDIS_ADDR"), Password: secretFromEnv("REDIS_PASSWORD"), DB: 0})
	defer redisClient.Close()
	if err = redisClient.Ping(ctx).Err(); err != nil {
		log.Fatal("Redis is unavailable")
	}
	repository := store.NewPostgres(db)
	limiter := usage.NewRedisFixedWindow(redisClient, 60, time.Minute)
	dlxClient := &translate.Client{BaseURL: dlxURL, InternalToken: requiredEnv("DLX_INTERNAL_TOKEN")}
	afdianClient := &afdian.Client{UserID: requiredEnv("AFDIAN_USER_ID"), Token: requiredEnv("AFDIAN_TOKEN")}
	server := &api.Server{Store: repository, Hasher: hasher, Sessions: sessions, Afdian: afdianClient, CookieDomain: os.Getenv("SESSION_COOKIE_DOMAIN"), DLX: dlxClient, Limiter: limiter, UpstreamSlots: make(chan struct{}, 32), MaxTextChars: 10000, ReadyCheck: func(ctx context.Context) error {
		if err := repository.Ping(ctx); err != nil {
			return err
		}
		if err := limiter.Ping(ctx); err != nil {
			return err
		}
		return dlxClient.Health(ctx)
	}}
	log.Printf("sponsor gateway listening on %s", address)
	if err := http.ListenAndServe(address, server.Routes()); err != nil {
		log.Fatal(err)
	}
}
