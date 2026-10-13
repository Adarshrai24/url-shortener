package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Adarshrai24/url-shortener/db"
	"github.com/Adarshrai24/url-shortener/models"
	"github.com/DATA-DOG/go-sqlmock"
)

func setupTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, func()) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	db.DB = mockDB
	os.Setenv("BASE_URL", "http://localhost:8090")
	return mockDB, mock, func() {
		mockDB.Close()
		db.DB = nil
	}
}

func makeTestServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{key}", GetURL)
	mux.HandleFunc("DELETE /{key}", DeleteURL)
	return httptest.NewServer(mux)
}

func TestShortenURL_NewURL(t *testing.T) {
	_, mock, cleanup := setupTestDB(t)
	defer cleanup()

	longURL := "https://example.com/very/long/url"

	mock.ExpectQuery(`SELECT id, key, short_url, long_url\s+FROM public\.url\s+WHERE long_url = \$1`).
		WithArgs(longURL).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectQuery(`SELECT EXISTS\(\s+SELECT 1 FROM public\.url WHERE key = \$1\s+\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectQuery(`INSERT INTO public\.url \(key, short_url, long_url\)\s+VALUES \(\$1, \$2, \$3\)\s+RETURNING id`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), longURL).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	reqBody := strings.NewReader(`{"url": "` + longURL + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/", reqBody)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	ShortenURL(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	var url models.Url
	if err := json.NewDecoder(rr.Body).Decode(&url); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if url.LongURL != longURL {
		t.Errorf("expected long_url %s, got %s", longURL, url.LongURL)
	}
	if url.ID != 1 {
		t.Errorf("expected id 1, got %d", url.ID)
	}
	if len(url.Key) != 7 {
		t.Errorf("expected key length 7, got %d", len(url.Key))
	}
	if !strings.HasPrefix(url.ShortURL, "http://localhost:8090/") {
		t.Errorf("expected short_url to start with base URL, got %s", url.ShortURL)
	}
}

func TestShortenURL_ExistingURL(t *testing.T) {
	_, mock, cleanup := setupTestDB(t)
	defer cleanup()

	longURL := "https://example.com/existing"
	key := "existing1"
	shortURL := "http://localhost:8090/" + key

	mock.ExpectQuery(`SELECT id, key, short_url, long_url\s+FROM public\.url\s+WHERE long_url = \$1`).
		WithArgs(longURL).
		WillReturnRows(sqlmock.NewRows([]string{"id", "key", "short_url", "long_url"}).AddRow(1, key, shortURL, longURL))

	reqBody := strings.NewReader(`{"url": "` + longURL + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/", reqBody)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	ShortenURL(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var url models.Url
	if err := json.NewDecoder(rr.Body).Decode(&url); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if url.LongURL != longURL {
		t.Errorf("expected long_url %s, got %s", longURL, url.LongURL)
	}
	if url.ShortURL != shortURL {
		t.Errorf("expected short_url %s, got %s", shortURL, url.ShortURL)
	}
}

func TestShortenURL_InvalidJSON(t *testing.T) {
	_, mock, cleanup := setupTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, key, short_url, long_url\s+FROM public\.url\s+WHERE long_url = \$1`).
		WithArgs("").
		WillReturnError(sql.ErrNoRows)

	mock.ExpectQuery(`SELECT EXISTS\(\s+SELECT 1 FROM public\.url WHERE key = \$1\s+\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectQuery(`INSERT INTO public\.url \(key, short_url, long_url\)\s+VALUES \(\$1, \$2, \$3\)\s+RETURNING id`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	reqBody := strings.NewReader(`{"invalid": "json"`)
	req := httptest.NewRequest(http.MethodPost, "/", reqBody)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	ShortenURL(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestShortenURL_MissingURL(t *testing.T) {
	_, mock, cleanup := setupTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, key, short_url, long_url\s+FROM public\.url\s+WHERE long_url = \$1`).
		WithArgs("").
		WillReturnError(sql.ErrNoRows)

	mock.ExpectQuery(`SELECT EXISTS\(\s+SELECT 1 FROM public\.url WHERE key = \$1\s+\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectQuery(`INSERT INTO public\.url \(key, short_url, long_url\)\s+VALUES \(\$1, \$2, \$3\)\s+RETURNING id`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	reqBody := strings.NewReader(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/", reqBody)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	ShortenURL(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}
}

func TestGetURL_Found(t *testing.T) {
	_, mock, cleanup := setupTestDB(t)
	defer cleanup()

	key := "testkey123"
	longURL := "https://example.com/found"

	mock.ExpectQuery(`SELECT id, key, short_url, long_url FROM public\.url`).
		WithArgs(key).
		WillReturnRows(sqlmock.NewRows([]string{"id", "key", "short_url", "long_url"}).AddRow(1, key, "http://localhost:8090/"+key, longURL))

	server := makeTestServer()
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/"+key, nil)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled mock expectations: %v", err)
	}

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected status %d, got %d", http.StatusFound, resp.StatusCode)
	}
	if resp.Header.Get("Location") != longURL {
		t.Errorf("expected redirect to %s, got %s", longURL, resp.Header.Get("Location"))
	}
}

func TestGetURL_NotFound(t *testing.T) {
	_, mock, cleanup := setupTestDB(t)
	defer cleanup()

	key := "nonexistent"

	mock.ExpectQuery(`SELECT id, key, short_url, long_url FROM public\.url`).
		WithArgs(key).
		WillReturnError(sql.ErrNoRows)

	server := makeTestServer()
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/"+key, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled mock expectations: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestDeleteURL_Success(t *testing.T) {
	_, mock, cleanup := setupTestDB(t)
	defer cleanup()

	key := "todelete"

	mock.ExpectExec(`DELETE FROM public\.url`).
		WithArgs(key).
		WillReturnResult(sqlmock.NewResult(0, 1))

	server := makeTestServer()
	defer server.Close()

	req, _ := http.NewRequest(http.MethodDelete, server.URL+"/"+key, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled mock expectations: %v", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, resp.StatusCode)
	}
}

func TestDeleteURL_NotFound(t *testing.T) {
	_, mock, cleanup := setupTestDB(t)
	defer cleanup()

	key := "nonexistent"

	mock.ExpectExec(`DELETE FROM public\.url`).
		WithArgs(key).
		WillReturnResult(sqlmock.NewResult(0, 0))

	server := makeTestServer()
	defer server.Close()

	req, _ := http.NewRequest(http.MethodDelete, server.URL+"/"+key, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled mock expectations: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestCandidateKey(t *testing.T) {
	_, mock, cleanup := setupTestDB(t)
	defer cleanup()

	longURL := "https://example.com/candidate"

	mock.ExpectQuery(`SELECT EXISTS\(\s+SELECT 1 FROM public\.url WHERE key = \$1\s+\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	key, err := candidateKey(longURL)
	if err != nil {
		t.Fatalf("candidateKey returned error: %v", err)
	}
	if len(key) != 7 {
		t.Errorf("expected key length 7, got %d", len(key))
	}
}

var client = &http.Client{}
