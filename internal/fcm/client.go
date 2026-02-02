package fcm

import (
    "context"
    "log"

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
    message := &messaging.MulticastMessage{
        Tokens: tokens,
        Notification: &messaging.Notification{
            Title: title,
            Body:  body,
        },
        Data: data,
    }
    _, err := client.sender.SendMulticast(ctx, message)
    return err
}
