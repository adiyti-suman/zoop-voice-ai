package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	verfRepo "verification-platform/internal/repository/verification"
	verfSvc "verification-platform/internal/service/verification"
	verfHandler "verification-platform/internal/handler/verification"
	appHttp "verification-platform/internal/http"
)

func healthHandler(db *sql.DB, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{
			"status":   "ok",
			"postgres": "ok",
			"redis":    "ok",
		}

		// Check Postgres
		if err := db.Ping(); err != nil {
			response["postgres"] = "error"
			response["status"] = "degraded"
		}

		// Check Redis
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := rdb.Ping(ctx).Err(); err != nil {
			response["redis"] = "error"
			response["status"] = "degraded"
		}

		w.Header().Set("Content-Type", "application/json")
		if response["status"] != "ok" {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		json.NewEncoder(w).Encode(response)
	}
}

func runMigrations(db *sql.DB) {
	// Quick hack for v0 instead of golang-migrate
	migration, err := os.ReadFile("./migrations/000001_create_verifications_table.up.sql")
	if err != nil {
		log.Printf("Could not read migration file: %v", err)
		return
	}
	_, err = db.Exec(string(migration))
	if err != nil {
		log.Printf("Migration 1 failed: %v", err)
	} else {
		log.Println("Migration 1 applied successfully.")
	}

	migration2, err2 := os.ReadFile("./migrations/000002_create_idempotency_keys_table.up.sql")
	if err2 == nil {
		_, err = db.Exec(string(migration2))
		if err != nil {
			log.Printf("Migration 2 failed: %v", err)
		} else {
			log.Println("Migration 2 applied successfully.")
		}
	}
}

func main() {
	// Setup Postgres Connection
	connStr := "postgres://postgres:password@localhost:5434/verification?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}
	defer db.Close()

	runMigrations(db)

	// Setup Redis Connection
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6380",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Wire Verification Domain
	repo := verfRepo.NewPostgresRepository(db)
	svc := verfSvc.NewService(repo)
	handler := verfHandler.NewHandler(svc)
	router := appHttp.NewRouter(handler)

	mux := http.NewServeMux()
	mux.Handle("/api/v1/", router)
	mux.HandleFunc("/health", healthHandler(db, rdb))

	port := ":8081"
	fmt.Printf("Server listening on port %s\n", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
