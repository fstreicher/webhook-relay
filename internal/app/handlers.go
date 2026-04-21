package app

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
)

const defaultWebhookTitle = "Webhook Alert"

type RelayEnvelope struct {
	Config  WebhookPayload `json:"config"`
	Payload interface{}    `json:"payload"`
}

// serviceHandler lists available relay services.
func serviceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	serviceNames := make([]string, 0, len(relays))
	for serviceName := range relays {
		serviceNames = append(serviceNames, serviceName)
	}
	sort.Strings(serviceNames)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"services": serviceNames,
	})
}

// serviceWebhookHandler routes incoming payloads to the selected service.
func serviceWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !authenticateRequest(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	serviceName := strings.TrimSpace(r.PathValue("service"))
	relay, exists := relays[serviceName]
	if !exists {
		http.Error(w, "Unknown service", http.StatusNotFound)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var envelope RelayEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if envelope.Config == nil {
		envelope.Config = WebhookPayload{}
	}

	payloadMap, _ := envelope.Payload.(map[string]interface{})

	title := defaultWebhookTitle
	if payloadMap != nil {
		title = extractString(payloadMap, "title", defaultWebhookTitle)
	}

	message := formatJSON(envelope.Payload)
	if payloadMap != nil {
		message = extractString(payloadMap, "message", message)
	}

	if err := relay(envelope.Config, envelope.Payload, title, message); err != nil {
		log.Printf("Failed to send to %s: %v", serviceName, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "accepted",
		"service": serviceName,
	})
}

// healthHandler is a simple endpoint to check if the server is running.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// extractString extracts a string value from the webhook payload.
func extractString(payload WebhookPayload, key, fallback string) string {
	if val, ok := payload[key]; ok {
		if str, ok := val.(string); ok && str != "" {
			return str
		}
	}
	return fallback
}

// formatJSON converts arbitrary JSON-compatible data to a formatted string.
func formatJSON(value interface{}) string {
	payloadBytes, _ := json.MarshalIndent(value, "", "  ")
	return string(payloadBytes)
}
