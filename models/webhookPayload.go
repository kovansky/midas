package models

import (
	"github.com/kovansky/strapi2hugo/models/enums"
	"time"
)

type WebhookPayload struct {
	Event     enums.StrapiWebhookEvents `json:"event"`
	CreatedAt time.Time                 `json:"createdAt"`
	Model     string                    `json:"model"`
	Entry     map[string]interface{}    `json:"entry"`
}
