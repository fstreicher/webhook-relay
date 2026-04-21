package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// sendToPushover sends the alert title and message to the Pushover service.
func sendToPushover(serviceConfig WebhookPayload, forwardedPayload interface{}, title, message string) error {
	_ = forwardedPayload

	appToken := strings.TrimSpace(extractString(serviceConfig, "token", ""))
	if appToken == "" {
		return fmt.Errorf("missing Pushover app token: set request field config.token")
	}

	userKey := strings.TrimSpace(extractString(serviceConfig, "user", ""))
	if userKey == "" {
		return fmt.Errorf("missing Pushover user key: set request field config.user")
	}

	params := url.Values{}
	params.Set("token", appToken)
	params.Set("user", userKey)
	params.Set("title", title)
	params.Set("message", message)

	if device := strings.TrimSpace(extractString(serviceConfig, "device", "")); device != "" {
		params.Set("device", device)
	}
	if sound := strings.TrimSpace(extractString(serviceConfig, "sound", "")); sound != "" {
		params.Set("sound", sound)
	}

	resp, err := http.PostForm("https://api.pushover.net/1/messages.json", params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if status, ok := result["status"].(float64); ok && status == 1 {
		return nil
	}
	if errorsList, ok := result["errors"].([]interface{}); ok && len(errorsList) > 0 {
		return fmt.Errorf("Pushover error: %v", errorsList[0])
	}
	return fmt.Errorf("Pushover error")
}
