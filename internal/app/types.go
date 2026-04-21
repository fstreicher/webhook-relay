package app

// WebhookPayload represents the JSON data received in webhook requests.
// It's a map where keys are strings and values can be any type.
type WebhookPayload map[string]interface{}

// Config holds all the application configuration settings.
type Config struct {
	Port          int             // The port number the server listens on.
	AllowedTokens map[string]bool // Map of valid authentication tokens.
}

// RelayFunc defines the function signature each webhook relay service must implement.
type RelayFunc func(serviceConfig WebhookPayload, forwardedPayload interface{}, title, message string) error
