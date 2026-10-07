package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/Adarshrai24/url-shortener/models"
	"github.com/Adarshrai24/url-shortener/utils"
	"github.com/Adarshrai24/url-shortener/db"
)

func ShortenURL(w http.ResponseWriter, r *http.Request) {
	longURL := r.URL.Query().Get("url")
	if longURL == "" {
		http.Error(w, "missing url parameter", http.StatusBadRequest)
		return
	}

	key := utils.Hash(longURL)

	var url models.Url
	err := db.DB.QueryRow(
		`SELECT id, key, short_url, long_url
		 FROM public.url
		 WHERE key = $1`,
		key,
	).Scan(&url.ID, &url.Key, &url.ShortURL, &url.LongURL)

	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(url)
		return
	}
	if err != sql.ErrNoRows {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	url = models.Url{
		LongURL:  longURL,
		ShortURL: "http://localhost:8090/" + key,
		Key:      key,
	}

	err = db.DB.QueryRow(
		`INSERT INTO public.url (key, short_url, long_url)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		url.Key, url.ShortURL, url.LongURL,
	).Scan(&url.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(url)
}


func GetURL(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	row := db.DB.QueryRow(
		`SELECT id, key, short_url, long_url FROM public.url
		where key = $1
		`,
		key,
	)
	
	var URL models.Url 
	err := row.Scan(
		&URL.ID, &URL.Key, &URL.ShortURL, &URL.LongURL,
	)
	
	if err == sql.ErrNoRows {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, URL.LongURL, http.StatusFound)
}