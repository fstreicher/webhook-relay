package app

import (
	"net/http"
	"strings"
)

// authenticateRequest checks if the incoming request has a valid authentication token.
func authenticateRequest(r *http.Request) bool {
	if len(config.AllowedTokens) == 0 {
		return true
	}

	auth := r.Header.Get("Authorization")
	if auth != "" {
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return config.AllowedTokens[parts[1]]
		}
	}

	return false
}
