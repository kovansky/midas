package models

type EventHandler interface {
	execute(payload WebhookPayload) (bool, error)
}
