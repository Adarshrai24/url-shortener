package main

import (
	"database/sql"
	"github.com/Adarshrai24/url-shortener/db"
	"github.com/Adarshrai24/url-shortener/handlers"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
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
	mux.HandleFunc("POST /", handlers.ShortenURL)
	mux.HandleFunc("GET /{key}", handlers.GetURL)
	mux.HandleFunc("DELETE /{key}", handlers.DeleteURL)
	http.ListenAndServe(":"+port, mux)
}
