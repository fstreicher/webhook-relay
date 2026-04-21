package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// sendToAlertzy sends the alert title and message to the Alertzy service.
func sendToAlertzy(serviceConfig WebhookPayload, forwardedPayload interface{}, title, message string) error {
	_ = forwardedPayload

	accountKey := strings.TrimSpace(extractString(serviceConfig, "accountKey", ""))
	if accountKey == "" {
		return fmt.Errorf("missing Alertzy account key: set request field config.accountKey")
	}

	group := strings.TrimSpace(extractString(serviceConfig, "group", ""))
	if group == "" {
		return fmt.Errorf("missing Alertzy group: set request field config.group")
	}

	params := url.Values{}
	params.Set("accountKey", accountKey)
	params.Set("title", title)
	params.Set("message", message)
	params.Set("group", group)

	// Optional parameters
	if priority := strings.TrimSpace(extractString(serviceConfig, "priority", "")); priority != "" {
		params.Set("priority", priority)
	}
	if image := strings.TrimSpace(extractString(serviceConfig, "image", "")); image != "" {
		params.Set("image", image)
	}
	if link := strings.TrimSpace(extractString(serviceConfig, "link", "")); link != "" {
		params.Set("link", link)
	}
	if buttons := strings.TrimSpace(extractString(serviceConfig, "buttons", "")); buttons != "" {
		params.Set("buttons", buttons)
	}

	resp, err := http.PostForm("https://alertzy.app/send", params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	// Response can be "success", "fail", or "mixed" (partial success across multiple account keys)
	response, _ := result["response"].(string)
	if response == "fail" || response == "mixed" {
		// error is a map of accountKey -> error message
		if errMap, ok := result["error"].(map[string]interface{}); ok {
			for key, msg := range errMap {
				return fmt.Errorf("Alertzy error for %s: %v", key, msg)
			}
		}
		return fmt.Errorf("Alertzy error (response: %s)", response)
	}

	return nil
}
