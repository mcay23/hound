package internal

import (
	"net"
	"sync"
	"time"
)

const (
	onlineCheckTTL     = 60 * time.Second
	onlineCheckTarget  = "8.8.8.8:53"
	onlineCheckTimeout = 3 * time.Second
)

type onlineChecker struct {
	mu         sync.Mutex
	checking   bool
	lastCheck  time.Time
	lastResult bool
}

var checker onlineChecker

// return immediately, refresh in background if necessary
func HasInternetConnection() bool {
	checker.mu.Lock()
	if time.Since(checker.lastCheck) < onlineCheckTTL {
		result := checker.lastResult
		checker.mu.Unlock()
		return result
	}
	if !checker.checking {
		checker.checking = true
		go checker.refresh()
	}
	result := checker.lastResult
	checker.mu.Unlock()

	return result
}

func (c *onlineChecker) refresh() {
	result := dial()
	c.mu.Lock()
	c.lastResult = result
	c.lastCheck = time.Now()
	c.checking = false
	c.mu.Unlock()
}

func dial() bool {
	conn, err := net.DialTimeout(
		"tcp",
		onlineCheckTarget,
		onlineCheckTimeout,
	)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
