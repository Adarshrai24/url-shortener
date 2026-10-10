package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Adarshrai24/url-shortener/db"
	"github.com/Adarshrai24/url-shortener/models"
	"github.com/Adarshrai24/url-shortener/utils"
)

func candidateKey(longURL string) (string, error) {
	for counter := 0; ; counter++ {
		input := longURL + ":" + strconv.Itoa(counter)
		key := utils.Hash(input)
		var exists bool
		err := db.DB.QueryRow(
			`SELECT EXISTS(
				SELECT 1 FROM public.url WHERE key = $1	
			)`,
			key,
		).Scan(&exists)
		if err != nil {
			return "", err
		}
		if !exists {
			return key, nil
		}
	}
}

func ShortenURL(w http.ResponseWriter, r *http.Request) {
	var req models.Req
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "missing url parameter", http.StatusBadRequest)
		return
	}
	longURL := req.URL	
	var url models.Url
	err = db.DB.QueryRow(
		`SELECT id, key, short_url, long_url
		 FROM public.url
		 WHERE long_url = $1`,
		longURL,
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

	key, err := candidateKey(longURL)
	if err != nil {
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
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, URL.LongURL, http.StatusFound)
}

func DeleteURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	key := r.PathValue("key")
	result, err := db.DB.Exec(
		`DELETE FROM public.url
		WHERE key = $1	
		`,
		key,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rows == 0 {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
