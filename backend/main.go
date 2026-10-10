package main

import (
	"log"
	"fmt"
	"net/http"
	"github.com/Adarshrai24/url-shortener/handlers"
	"github.com/Adarshrai24/url-shortener/db"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
	"os"
)

func main(){
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	user := os.Getenv("DB_USER")
	host := os.Getenv("DB_HOST")
	password := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")
	connStr := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s sslmode=disable",
		host, user, password, name,
	)
	conn, err := sql.Open("postgres", connStr)
	db.DB = conn
	if err != nil {
		panic(err)
	}
	err = conn.Ping()
	if err != nil {
		log.Fatal(err)
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
	http.ListenAndServe(":8090", mux)
}