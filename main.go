// Package main is the entry point for the webhook relay application
package main

import (
	"encoding/json" // For JSON encoding/decoding
	"flag"          // For command-line flags
	"fmt"           // For formatted string output
	"io"            // For I/O operations like reading request bodies
	"log"           // For logging messages
	"net/http"      // For HTTP server and client functionality
	"net/url"       // For URL query parameter handling
	"os"            // For operating system operations (env vars, etc)
	"strconv"       // For string-to-number conversions
	"strings"       // For string manipulation functions
)

// WebhookPayload represents the JSON data received in webhook requests
// It's a map where keys are strings and values can be any type
type WebhookPayload map[string]interface{}

// Config holds all the application configuration settings
type Config struct {
	Port          int             // The port number the server listens on
	AlertzyKey    string          // API key for Alertzy service
	AlertzyGroup  string          // Alert group name in Alertzy
	DefaultTitle  string          // Default title if webhook doesn't provide one
	AllowedTokens map[string]bool // Map of valid authentication tokens
}

// Global config variable holds the application configuration
var config Config

// init() runs automatically when the program starts, before main()
// It sets up configuration from command-line flags and environment variables
func init() {
	port := flag.Int("port", 8080, "Port to listen on")
	alertzyKey := flag.String("key", "", "Alertzy account key (or set ALERTZY_KEY env var)")
	alertzyGroup := flag.String("group", "webhooks", "Alertzy group name")
	defaultTitle := flag.String("title", "Webhook Alert", "Default alert title")
	token := flag.String("token", "", "Optional auth token (or set WEBHOOK_TOKEN env var)")
	flag.Parse() // Parse command-line flags

	// If no key was provided via flag, try to get it from environment variable
	if *alertzyKey == "" {
		*alertzyKey = os.Getenv("ALERTZY_KEY")
	}
	// If still no key, the app cannot run, so exit with error
	if *alertzyKey == "" {
		log.Fatal("Alertzy key required: set -key flag or ALERTZY_KEY environment variable")
	}

	// Initialize the config struct with the parsed and validated values
	config = Config{
		Port:          *port,
		AlertzyKey:    *alertzyKey,
		AlertzyGroup:  *alertzyGroup,
		DefaultTitle:  *defaultTitle,
		AllowedTokens: make(map[string]bool), // Create an empty map for tokens
	}

	// Add the single token from flag if provided
	if *token != "" {
		config.AllowedTokens[*token] = true
	}
	// Add tokens from environment variable (comma-separated list)
	if envToken := os.Getenv("WEBHOOK_TOKEN"); envToken != "" {
		for _, token := range strings.Split(envToken, ",") {
			if token = strings.TrimSpace(token); token != "" { // Trim whitespace from each token
				config.AllowedTokens[token] = true
			}
		}
	}
}

// authenticateRequest checks if the incoming request has a valid authentication token
// Returns true if the request is authorized, false otherwise
func authenticateRequest(r *http.Request) bool {
	// If no tokens are configured, allow all requests
	if len(config.AllowedTokens) == 0 {
		return true
	}

	// Try to get token from Authorization header (e.g., "Bearer mytoken123")
	auth := r.Header.Get("Authorization")
	if auth != "" {
		parts := strings.SplitN(auth, " ", 2) // Split into ["Bearer", "token"]
		if len(parts) == 2 && parts[0] == "Bearer" {
			// Check if this token is in our allowed list
			return config.AllowedTokens[parts[1]]
		}
	}

	// No valid token found
	return false
}

// webhookHandler processes incoming webhook requests
// w is the response writer, r is the incoming request
func webhookHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if the request has valid authentication
	if !authenticateRequest(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Read the request body (the raw data sent to us)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close() // Ensure the body is closed after reading

	// Parse the JSON data into our WebhookPayload map
	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Extract title and message from the payload, with fallback values
	title := extractString(payload, "title", config.DefaultTitle)
	message := extractString(payload, "message", formatPayload(payload))

	// Send the alert to Alertzy service
	if err := sendToAlertzy(title, message); err != nil {
		log.Printf("Failed to send to Alertzy: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Send a successful response back to the client
	w.Header().Set("Content-Type", "application/json")                 // Tell client we're sending JSON
	w.WriteHeader(http.StatusAccepted)                                 // HTTP 202: request accepted
	json.NewEncoder(w).Encode(map[string]string{"status": "accepted"}) // Send response
}

// extractString extracts a string value from the webhook payload
// Returns the value if found, otherwise returns the fallback string
func extractString(payload WebhookPayload, key, fallback string) string {
	if val, ok := payload[key]; ok { // Check if key exists in payload
		if str, ok := val.(string); ok && str != "" { // Assert it's a non-empty string
			return str
		}
	}
	return fallback // No valid value found, use fallback
}

// formatPayload converts the entire webhook payload to a nicely formatted JSON string
func formatPayload(payload WebhookPayload) string {
	bytes, _ := json.MarshalIndent(payload, "", "  ") // Marshal with 2-space indentation
	return string(bytes)                              // Convert bytes to string
}

// sendToAlertzy sends the alert title and message to the Alertzy service
func sendToAlertzy(title, message string) error {
	// Build the query parameters for the Alertzy API request
	params := url.Values{}
	params.Set("accountKey", config.AlertzyKey)
	params.Set("title", title)
	params.Set("message", message)
	params.Set("group", config.AlertzyGroup)

	// Make the HTTP POST request to Alertzy
	resp, err := http.PostForm("https://alertzy.app/send", params)
	if err != nil {
		return err // Network error
	}
	defer resp.Body.Close() // Ensure response body is closed

	// Read and parse the response from Alertzy
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	// Check if Alertzy returned an error
	if response, ok := result["response"].(string); ok && response == "fail" {
		if errMsg, ok := result["error"].(string); ok {
			return fmt.Errorf("Alertzy error: %s", errMsg)
		}
		return fmt.Errorf("Alertzy error")
	}

	return nil // Success
}

// healthHandler is a simple endpoint to check if the server is running
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// main is the entry point of the application
func main() {
	// Register URL routes and their handler functions
	http.HandleFunc("/health", healthHandler)   // Health check endpoint
	http.HandleFunc("/webhook", webhookHandler) // Webhook endpoint
	http.HandleFunc("/", webhookHandler)        // Default endpoint (also handles webhooks)

	// Convert port number to string and create the address
	addr := ":" + strconv.Itoa(config.Port)
	log.Printf("Starting webhook relay on %s", addr)
	log.Printf("Alertzy group: %s", config.AlertzyGroup)
	if len(config.AllowedTokens) > 0 {
		log.Printf("Authentication enabled")
	}

	// Start the HTTP server (this blocks until the server exits)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
