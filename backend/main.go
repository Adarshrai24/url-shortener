package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/Adarshrai24/url-shortener/db"
	"github.com/Adarshrai24/url-shortener/handlers"
	"github.com/Adarshrai24/url-shortener/middleware"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	_ = godotenv.Load()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	conn, err := sql.Open("postgres", databaseURL)
	db.DB = conn
	if err != nil {
		panic(err)
	}
	err = conn.Ping()
	if err != nil {
		log.Fatal(err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}
	mux := http.NewServeMux()
	mux.Handle(
		"GET /static/",
		http.StripPrefix(
			"/static/",
			http.FileServer(http.Dir("../frontend")),
		),
	)
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/index.html")
	})
	mux.Handle("POST /", middleware.RateLimiter(
		http.HandlerFunc(handlers.ShortenURL),
	))
	mux.Handle("GET /{key}", middleware.RateLimiter(
		http.HandlerFunc(handlers.GetURL),
	))
	http.ListenAndServe(":"+port, mux)
}
