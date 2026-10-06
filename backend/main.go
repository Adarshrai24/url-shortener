package main

import (
	"net/http"

	"github.com/Adarshrai24/url-shortener/handlers"
)

func main(){
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", handlers.ShortenURL)
	http.ListenAndServe(":8090", mux)
}