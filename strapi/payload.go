package strapi

import (
	"encoding/json"
	"github.com/kovansky/strapi2hugo"
	"time"
)

var _ strapi2hugo.Payload = (*Payload)(nil)

type Payload struct {
	event     event
	CreatedAt time.Time
	Model     string
	entry     map[string]interface{}
}

func ParsePayload(json []byte) (strapi2hugo.Payload, error) {
	payload := Payload{}
	err := payload.UnmarshalJSON(json)

	if err != nil {
		return nil, err
	}

	return payload, nil
}

func (p Payload) Event() string {
	return p.event.String()
}

func (p Payload) Metadata() map[string]interface{} {
	asMap := make(map[string]interface{})

	asMap["event"] = p.event
	asMap["createdAt"] = p.CreatedAt
	asMap["model"] = p.Model

	return asMap
}

func (p Payload) Entry() map[string]interface{} {
	return p.entry
}

func (p Payload) Raw() interface{} {
	return p
}

func (p Payload) MarshalJSON() ([]byte, error) {
	j, err := json.Marshal(struct {
		Event     event                  `json:"event"`
		CreatedAt time.Time              `json:"createdAt"`
		Model     string                 `json:"model"`
		Entry     map[string]interface{} `json:"entry,omitempty"`
	}{
		p.event, p.CreatedAt, p.Model, p.entry,
	})

	if err != nil {
		return nil, err
	}

	return j, nil
}

func (p *Payload) UnmarshalJSON(bytes []byte) error {
	var data map[string]interface{}
	err := json.Unmarshal(bytes, &data)
	if err != nil {
		return err
	}

	newPayload := Payload{
		event:     data["event"].(event),
		CreatedAt: data["createdAt"].(time.Time),
		Model:     data["model"].(string),
		entry:     data["entry"].(map[string]interface{}),
	}

	*p = newPayload
	return nil
}
