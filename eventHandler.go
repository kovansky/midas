package main

type EventHandler interface {
    execute(payload WebhookPayload) (bool, error)
}
