package corereader2

import (
	"net/http"
	"time"
)

// Advised by Gemini AI to avoid "Only one usage of each socket address (...) is normally permitted"

var TheOneAndOnlyTransport = &http.Transport{
	// High numbers to saturate my 36 cores
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 100,
	// Keeps connections open so they don't go into TIME_WAIT
	IdleConnTimeout: 90 * time.Second,
}

var TheOneAndOnlyClient = &http.Client{
	Transport: TheOneAndOnlyTransport,
}
