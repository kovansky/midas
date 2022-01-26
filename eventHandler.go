package strapi2hugo

type EventHandler interface {
	execute(payload WebhookPayload) (bool, error)
}
