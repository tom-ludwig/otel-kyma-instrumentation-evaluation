package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var db *sql.DB

func main() {
	var err error
	dsn := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		getenv("PG_HOST", "postgres"),
		getenv("PG_PORT", "5432"),
		getenv("PG_DB", "postgres"),
		getenv("PG_USER", "postgres"),
		getenv("PG_PASSWORD", "postgres"),
	)
	db, err = sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer db.Close()

	http.HandleFunc("/rolldice", rolldice)
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func rolldice(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)

	if _, err := db.ExecContext(r.Context(), "SELECT 1"); err != nil {
		log.Printf("db query failed: %v", err)
		http.Error(w, fmt.Sprintf("db error: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "rolled: %d\n", rand.Intn(6)+1)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
