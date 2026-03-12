package fcm

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/url"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type Client struct {
	enabled bool
	sender  *messaging.Client
	webApp  string
}

func NewClient(enabled bool, credentialsPath string, webAppURL string) *Client {
	if !enabled {
		return &Client{enabled: false, webApp: normalizeWebAppURL(webAppURL)}
	}
	credentialsOption, err := resolveCredentialOption(credentialsPath)
	if err != nil {
		log.Printf("FCM disabled: invalid credentials format: %v", err)
		return &Client{enabled: false, webApp: normalizeWebAppURL(webAppURL)}
	}
	app, err := firebase.NewApp(context.Background(), nil, credentialsOption)
	if err != nil {
		log.Printf("FCM disabled: failed to init app: %v", err)
		return &Client{enabled: false, webApp: normalizeWebAppURL(webAppURL)}
	}
	sender, err := app.Messaging(context.Background())
	if err != nil {
		log.Printf("FCM disabled: failed to init messaging: %v", err)
		return &Client{enabled: false, webApp: normalizeWebAppURL(webAppURL)}
	}
	return &Client{
		enabled: true,
		sender:  sender,
		webApp:  normalizeWebAppURL(webAppURL),
	}
}

func resolveCredentialOption(raw string) (option.ClientOption, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, errors.New("empty credentials")
	}

	// Raw service account JSON.
	if strings.HasPrefix(trimmed, "{") && json.Valid([]byte(trimmed)) {
		return option.WithCredentialsJSON([]byte(trimmed)), nil
	}

	// Base64 encoded service account JSON (recommended for Heroku config vars).
	if decoded, err := base64.StdEncoding.DecodeString(trimmed); err == nil && json.Valid(decoded) {
		return option.WithCredentialsJSON(decoded), nil
	}

	// Fallback to filesystem path (local/dev usage).
	return option.WithCredentialsFile(trimmed), nil
}

func (client *Client) SendToTokens(
	ctx context.Context,
	tokens []string,
	title string,
	body string,
	data map[string]string,
) ([]string, error) {
	if !client.enabled || client.sender == nil || len(tokens) == 0 {
		return nil, nil
	}
	successCount := 0
	failureCount := 0
	invalidCount := 0
	invalidTokens := make([]string, 0)
	var lastErr error

	for index, token := range tokens {
		if strings.TrimSpace(token) == "" {
			continue
		}
		webpushData := copyStringMap(data)
		if webpushData == nil {
			webpushData = make(map[string]string)
		}
		if _, ok := webpushData["title"]; !ok && strings.TrimSpace(title) != "" {
			webpushData["title"] = title
		}
		if _, ok := webpushData["body"]; !ok && strings.TrimSpace(body) != "" {
			webpushData["body"] = body
		}
		webpush := &messaging.WebpushConfig{
			Headers: map[string]string{
				"Urgency": "high",
			},
			Data: webpushData,
		}
		if link := client.webNotificationLink(webpushData); link != "" {
			webpush.FCMOptions = &messaging.WebpushFCMOptions{
				Link: link,
			}
			webpush.Data["link"] = link
		}
		message := &messaging.Message{
			Token: token,
			Data:  copyStringMap(data),
			Android: &messaging.AndroidConfig{
				Priority: "high",
				Notification: &messaging.AndroidNotification{
					Title:        title,
					Body:         body,
					ChannelID:    "helpdesk_updates",
					Priority:     messaging.PriorityMax,
					DefaultSound: true,
				},
			},
			APNS: &messaging.APNSConfig{
				Headers: map[string]string{
					"apns-priority":  "10",
					"apns-push-type": "alert",
				},
				Payload: &messaging.APNSPayload{
					Aps: &messaging.Aps{
						Alert: &messaging.ApsAlert{
							Title: title,
							Body:  body,
						},
						Sound: "default",
					},
					CustomData: copyInterfaceMap(data),
				},
			},
			Webpush: webpush,
		}
		if _, err := client.sender.Send(ctx, message); err != nil {
			tokenHint := token
			if len(tokenHint) > 10 {
				tokenHint = tokenHint[:10] + "..."
			}
			if isInvalidTokenError(err) {
				invalidCount++
				invalidTokens = append(invalidTokens, token)
				log.Printf("fcm invalid token[%d]=%s: %v", index, tokenHint, err)
				continue
			}
			failureCount++
			lastErr = err
			log.Printf("fcm send failed token[%d]=%s: %v", index, tokenHint, err)
			continue
		}
		successCount++
	}

	log.Printf("fcm sent: success=%d invalid=%d failure=%d", successCount, invalidCount, failureCount)
	return invalidTokens, lastErr
}

func (client *Client) webNotificationLink(data map[string]string) string {
	if client.webApp == "" {
		return ""
	}

	target, err := url.Parse(client.webApp)
	if err != nil {
		return ""
	}

	target.Path = "/"
	target.RawQuery = ""
	target.Fragment = "/notifications"
	if ticketID := strings.TrimSpace(data["ticket_id"]); ticketID != "" {
		target.Fragment += "?ticketId=" + url.QueryEscape(ticketID)
	}
	return target.String()
}

func normalizeWebAppURL(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	return strings.TrimRight(trimmed, "/")
}

func copyStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	result := make(map[string]string, len(input))
	for key, value := range input {
		result[key] = value
	}
	return result
}

func copyInterfaceMap(input map[string]string) map[string]interface{} {
	if len(input) == 0 {
		return nil
	}
	result := make(map[string]interface{}, len(input))
	for key, value := range input {
		result[key] = value
	}
	return result
}

func isInvalidTokenError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "requested entity was not found") ||
		strings.Contains(message, "registration-token-not-registered") ||
		strings.Contains(message, "unregistered")
}
