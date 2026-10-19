package utils

import (
	"testing"
	"time"
)

func resetTokenBucket(t *testing.T, ipAddress string) {
	t.Helper()

	mu.Lock()
	delete(tokenCounter, ipAddress)
	mu.Unlock()

	t.Cleanup(func() {
		mu.Lock()
		delete(tokenCounter, ipAddress)
		mu.Unlock()
	})
}

func TestTokenBucketAllowsBurstAndRejectsExcess(t *testing.T) {
	const ipAddress = "test-burst"
	resetTokenBucket(t, ipAddress)

	for request := 1; request <= maxToken; request++ {
		if !TokenBucket(ipAddress) {
			t.Fatalf("TokenBucket() rejected request %d of %d", request, maxToken)
		}
	}

	if TokenBucket(ipAddress) {
		t.Fatal("TokenBucket() allowed a request after the bucket was exhausted")
	}
}

func TestTokenBucketRefillsOverTime(t *testing.T) {
	const ipAddress = "test-refill"
	resetTokenBucket(t, ipAddress)

	mu.Lock()
	tokenCounter[ipAddress] = ClientLimiter{
		tokens:     0,
		lastRefill: time.Now().Add(-2 * time.Second),
	}
	mu.Unlock()

	if !TokenBucket(ipAddress) {
		t.Fatal("TokenBucket() rejected a request despite elapsed refill time")
	}

	mu.Lock()
	remainingTokens := tokenCounter[ipAddress].tokens
	mu.Unlock()
	if remainingTokens < 1 || remainingTokens >= 2 {
		t.Errorf("remaining tokens = %v, want at least 1 and less than 2", remainingTokens)
	}
}

func TestTokenBucketCapsRefillAtMaximum(t *testing.T) {
	const ipAddress = "test-refill-cap"
	resetTokenBucket(t, ipAddress)

	mu.Lock()
	tokenCounter[ipAddress] = ClientLimiter{
		tokens:     0,
		lastRefill: time.Now().Add(-2 * time.Minute),
	}
	mu.Unlock()

	if !TokenBucket(ipAddress) {
		t.Fatal("TokenBucket() rejected a request after the bucket refilled")
	}

	mu.Lock()
	remainingTokens := tokenCounter[ipAddress].tokens
	mu.Unlock()
	if remainingTokens < maxToken-1 || remainingTokens >= maxToken {
		t.Errorf("remaining tokens = %v, want %v after consuming from a full bucket", remainingTokens, maxToken-1)
	}
}

func TestTokenBucketTracksClientsIndependently(t *testing.T) {
	const firstIP = "test-client-one"
	const secondIP = "test-client-two"
	resetTokenBucket(t, firstIP)
	resetTokenBucket(t, secondIP)

	for request := 1; request <= maxToken; request++ {
		if !TokenBucket(firstIP) {
			t.Fatalf("TokenBucket() rejected first client's request %d", request)
		}
	}

	if !TokenBucket(secondIP) {
		t.Fatal("TokenBucket() rejected a request from a different client")
	}
	if TokenBucket(firstIP) {
		t.Fatal("TokenBucket() allowed an extra request from the exhausted first client")
	}
}
