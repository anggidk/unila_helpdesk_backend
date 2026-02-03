package fcm

import (
    "context"
    "log"
    "strings"

    firebase "firebase.google.com/go/v4"
    "firebase.google.com/go/v4/messaging"
    "google.golang.org/api/option"
)

type Client struct {
    enabled bool
    sender  *messaging.Client
}

func NewClient(enabled bool, credentialsPath string) *Client {
    if !enabled {
        return &Client{enabled: false}
    }
    app, err := firebase.NewApp(context.Background(), nil, option.WithCredentialsFile(credentialsPath))
    if err != nil {
        log.Printf("FCM disabled: failed to init app: %v", err)
        return &Client{enabled: false}
    }
    sender, err := app.Messaging(context.Background())
    if err != nil {
        log.Printf("FCM disabled: failed to init messaging: %v", err)
        return &Client{enabled: false}
    }
    return &Client{enabled: true, sender: sender}
}

func (client *Client) SendToTokens(ctx context.Context, tokens []string, title string, body string, data map[string]string) error {
    if !client.enabled || client.sender == nil || len(tokens) == 0 {
        return nil
    }
    successCount := 0
    failureCount := 0
    var lastErr error

    for index, token := range tokens {
        if strings.TrimSpace(token) == "" {
            continue
        }
        message := &messaging.Message{
            Token: token,
            Notification: &messaging.Notification{
                Title: title,
                Body:  body,
            },
            Data: data,
        }
        if _, err := client.sender.Send(ctx, message); err != nil {
            failureCount++
            lastErr = err
            tokenHint := token
            if len(tokenHint) > 10 {
                tokenHint = tokenHint[:10] + "..."
            }
            log.Printf("fcm send failed token[%d]=%s: %v", index, tokenHint, err)
            continue
        }
        successCount++
    }

    log.Printf("fcm sent: success=%d failure=%d", successCount, failureCount)
    return lastErr
}
