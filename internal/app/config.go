package app

import (
	"flag"
	"os"
	"strings"
)

// Global config variable holds the application configuration.
var config Config

// init runs automatically before main and sets up app config.
func init() {
	port := flag.Int("port", 8080, "Port to listen on")
	token := flag.String("token", "", "Optional auth token (or set WEBHOOK_TOKEN env var)")
	flag.Parse()

	config = Config{
		Port:          *port,
		AllowedTokens: make(map[string]bool),
	}

	if *token != "" {
		config.AllowedTokens[*token] = true
	}
	if envToken := os.Getenv("WEBHOOK_TOKEN"); envToken != "" {
		for _, token := range strings.Split(envToken, ",") {
			if token = strings.TrimSpace(token); token != "" {
				config.AllowedTokens[token] = true
			}
		}
	}
}
