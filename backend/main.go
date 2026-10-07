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
	var database, schema string
	err = conn.QueryRow(
		`SELECT current_database(), current_schema()`,
	).Scan(&database, &schema)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Connected to database=%q schema=%q", database, schema)
	var tableName *string

	err = conn.QueryRow(`SELECT to_regclass('public.url')`).Scan(&tableName)
	if err != nil {
		log.Fatal(err)
	}

	if tableName == nil {
		log.Println("Table public.url does not exist")
	} else {
		log.Printf("Table exists: %s", *tableName)
	}


	mux := http.NewServeMux()
	mux.HandleFunc("POST /", handlers.ShortenURL)
	mux.HandleFunc("GET /{key}", handlers.GetURL)
	http.ListenAndServe(":8090", mux)
}