package app

import (
	"log"
	"net/http"
	"sort"
	"strconv"
)

// Run configures routes and starts the HTTP server.
func Run() {
	http.HandleFunc("GET /health", healthHandler)
	http.HandleFunc("GET /services", serviceHandler)
	http.HandleFunc("POST /{service}", serviceWebhookHandler)

	addr := ":" + strconv.Itoa(config.Port)
	log.Printf("Starting webhook relay on %s", addr)
	if len(config.AllowedTokens) > 0 {
		log.Printf("Authentication enabled")
	}

	log.Printf("[RouterExplorer] Mapped {/health, GET} route")
	log.Printf("[RouterExplorer] Mapped {/services, GET} route")

	names := make([]string, 0, len(relays))
	for name := range relays {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		log.Printf("[RouterExplorer] Mapped {/%s, POST} route", name)
	}

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
