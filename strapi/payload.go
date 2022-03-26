/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package strapi

import (
	"encoding/json"
	"github.com/kovansky/midas"
	"time"
)

var _ midas.Payload = (*Payload)(nil)

type Payload struct {
	event     Event
	CreatedAt time.Time
	Model     string
	metadata  map[string]interface{}
	entry     map[string]interface{}
}

func ParsePayload(json []byte) (midas.Payload, error) {
	payload := Payload{}
	err := payload.UnmarshalJSON(json)

	if err != nil {
		return nil, err
	}

	payload.createMetadataMap()

	return &payload, nil
}

func (p Payload) Event() string {
	return p.event.String()
}

func (p Payload) Metadata() map[string]interface{} {
	return p.metadata
}

func (p *Payload) createMetadataMap() {
	asMap := make(map[string]interface{})

	asMap["event"] = p.event
	asMap["createdAt"] = p.CreatedAt
	asMap["model"] = p.Model

	asMap["published"] = false
	if val, ok := p.entry["publishedAt"]; ok {
		asMap["published"] = val != nil
	}

	p.metadata = asMap
}

func (p Payload) Entry() map[string]interface{} {
	return p.entry
}

func (p Payload) Raw() interface{} {
	return p
}

func (p Payload) MarshalJSON() ([]byte, error) {
	j, err := json.Marshal(struct {
		Event     Event                  `json:"event"`
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
	var data struct {
		Event     Event                  `json:"event"`
		CreatedAt time.Time              `json:"createdAt"`
		Model     string                 `json:"model"`
		Entry     map[string]interface{} `json:"entry,omitempty"`
	}
	err := json.Unmarshal(bytes, &data)
	if err != nil {
		return err
	}

	newPayload := Payload{
		event:     data.Event,
		CreatedAt: data.CreatedAt,
		Model:     data.Model,
		entry:     data.Entry,
	}

	*p = newPayload
	return nil
}
