package utils

import (
	"sync"
	"time"
)

type ClientLimiter struct {
	tokens     float64
	lastRefill time.Time
}

var (
	tokenCounter = make(map[string]ClientLimiter)
	mu           sync.Mutex
)

const maxToken = 60

func TokenBucket(ipAddress string) bool {
	mu.Lock()
	defer mu.Unlock()
	value, present := tokenCounter[ipAddress]
	if !present {
		currentTime := time.Now()
		tokenCounter[ipAddress] = ClientLimiter{maxToken - 1, currentTime}
		return true
	} else {
		tokenCount := value.tokens
		currentTime := time.Now()
		elapasedTime := currentTime.Sub(value.lastRefill)
		totalTokenToRefill := min(tokenCount+elapasedTime.Seconds(), 60)
		if totalTokenToRefill < 1 {
			tokenCounter[ipAddress] = ClientLimiter{totalTokenToRefill, currentTime}
			return false
		}
		tokenCounter[ipAddress] = ClientLimiter{totalTokenToRefill - 1, currentTime}
		return true
	}
}
