// Package holded contains the upstream API client, config, and context plumbing.
package holded

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://api.holded.com/api/invoicing/v1"

	apiKeyEnvVar           = "HOLDED_API_KEY"
	apiBaseEnvVar          = "HOLDED_API_BASE"
	timeoutMsEnvVar        = "HOLDED_TIMEOUT_MS"
	allowWriteEnvVar       = "HOLDED_ALLOW_WRITE"
	allowedToolsEnvVar     = "HOLDED_ALLOWED_TOOLS"
	debugEnvVar            = "HOLDED_DEBUG"
	disableRateLimitEnvVar = "HOLDED_RATE_LIMIT_DISABLE"

	urlHeader    = "X-HOLDED-URL"
	apiKeyHeader = "X-HOLDED-API-Key"
)

type Config struct {
	Debug                   bool
	IncludeArgumentsInSpans bool
	URL                     string
	APIKey                  string
	AllowWrite              bool
	AllowedTools            map[string]struct{}
	Timeout                 time.Duration
	DisableRateLimit        bool
}

func envBool(key string) bool {
	val := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	return val == "1" || val == "true" || val == "yes"
}

func allowedToolsFromEnv() map[string]struct{} {
	val := strings.TrimSpace(os.Getenv(allowedToolsEnvVar))
	if val == "" {
		return nil
	}
	set := make(map[string]struct{})
	for _, t := range strings.Split(val, ",") {
		name := strings.TrimSpace(t)
		if name != "" {
			set[name] = struct{}{}
		}
	}
	return set
}

func timeoutFromEnv() time.Duration {
	val := strings.TrimSpace(os.Getenv(timeoutMsEnvVar))
	if val == "" {
		return 30 * time.Second
	}
	ms, err := strconv.Atoi(val)
	if err != nil || ms <= 0 {
		return 30 * time.Second
	}
	return time.Duration(ms) * time.Millisecond
}

func baseURLFromEnv() string {
	u := strings.TrimRight(os.Getenv(apiBaseEnvVar), "/")
	if u == "" {
		return defaultBaseURL
	}
	return u
}

func apiKeyFromEnv() string {
	return strings.TrimSpace(os.Getenv(apiKeyEnvVar))
}
