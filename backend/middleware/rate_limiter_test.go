package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimiterPassesAllowedRequestToNextHandler(t *testing.T) {
	nextCalled := false
	handler := RateLimiter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.20:12345"
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Errorf("status code = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if !nextCalled {
		t.Error("next handler was not called for an allowed request")
	}
}

func TestRateLimiterRejectsInvalidClientAddress(t *testing.T) {
	nextCalled := false
	handler := RateLimiter(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextCalled = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "invalid-address"
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if nextCalled {
		t.Error("next handler was called for an invalid client address")
	}
}

func TestRateLimiterRejectsRequestAfterLimit(t *testing.T) {
	const clientAddress = "192.0.2.21:12345"
	handler := RateLimiter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for request := 1; request <= 60; request++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = clientAddress
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusNoContent {
			t.Fatalf("request %d status code = %d, want %d", request, recorder.Code, http.StatusNoContent)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = clientAddress
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusTooManyRequests {
		t.Errorf("status code = %d, want %d", recorder.Code, http.StatusTooManyRequests)
	}
}
