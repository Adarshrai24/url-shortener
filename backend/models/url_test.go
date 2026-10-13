package models

import (
	"encoding/json"
	"testing"
)

func TestReq_JSONMarshal(t *testing.T) {
	req := Req{URL: "https://example.com"}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Req
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.URL != req.URL {
		t.Errorf("expected URL %s, got %s", req.URL, decoded.URL)
	}
}

func TestReq_JSONUnmarshal(t *testing.T) {
	data := []byte(`{"url": "https://example.com"}`)
	var req Req
	if err := json.Unmarshal(data, &req); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if req.URL != "https://example.com" {
		t.Errorf("expected URL https://example.com, got %s", req.URL)
	}
}

func TestURL_JSONMarshal(t *testing.T) {
	url := Url{
		ID:       1,
		Key:      "abc1234",
		ShortURL: "http://localhost:8090/abc1234",
		LongURL:  "https://example.com",
	}
	data, err := json.Marshal(url)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Url
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.ID != url.ID {
		t.Errorf("expected ID %d, got %d", url.ID, decoded.ID)
	}
	if decoded.Key != url.Key {
		t.Errorf("expected Key %s, got %s", url.Key, decoded.Key)
	}
	if decoded.ShortURL != url.ShortURL {
		t.Errorf("expected ShortURL %s, got %s", url.ShortURL, decoded.ShortURL)
	}
	if decoded.LongURL != url.LongURL {
		t.Errorf("expected LongURL %s, got %s", url.LongURL, decoded.LongURL)
	}
}

func TestURL_JSONUnmarshal(t *testing.T) {
	data := []byte(`{"id": 1, "key": "abc1234", "short_url": "http://localhost:8090/abc1234", "long_url": "https://example.com"}`)
	var url Url
	if err := json.Unmarshal(data, &url); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if url.ID != 1 {
		t.Errorf("expected ID 1, got %d", url.ID)
	}
	if url.Key != "abc1234" {
		t.Errorf("expected Key abc1234, got %s", url.Key)
	}
	if url.ShortURL != "http://localhost:8090/abc1234" {
		t.Errorf("expected ShortURL http://localhost:8090/abc1234, got %s", url.ShortURL)
	}
	if url.LongURL != "https://example.com" {
		t.Errorf("expected LongURL https://example.com, got %s", url.LongURL)
	}
}
