package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Adarshrai24/url-shortener/models"
	"github.com/Adarshrai24/url-shortener/utils"
)

func ShortenURL(w http.ResponseWriter, r *http.Request) {
	longURL := r.URL.Query().Get("url")	
	
	key := utils.Hash(longURL)
	shortURL := "http://localhost:8090/" + key 
	url := models.Url {
		LongURL: longURL,
		ShortURL: shortURL,
		Key: key,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err := json.NewEncoder(w).Encode(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}