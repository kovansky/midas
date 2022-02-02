package midas

import (
	"bytes"
	"encoding/json"
)

type StrapiWebhookEvents int64

const (
	Undefined StrapiWebhookEvents = iota
	Create
	Update
	Delete
	Publish
	Unpublish
)

var toJson = map[StrapiWebhookEvents]string{
	Undefined: "",
	Create:    "entry.create",
	Update:    "entry.update",
	Delete:    "entry.delete",
	Publish:   "entry.publish",
	Unpublish: "entry.unpublish",
}

var toString = map[StrapiWebhookEvents]string{
	Undefined: "",
	Create:    "Create",
	Update:    "Update",
	Delete:    "Delete",
	Publish:   "Publish",
	Unpublish: "Unpublish",
}

var toId = map[string]StrapiWebhookEvents{
	"entry.create":    Create,
	"entry.update":    Update,
	"entry.delete":    Delete,
	"entry.publish":   Publish,
	"entry.unpublish": Unpublish,
}

func (event StrapiWebhookEvents) String() string {
	return toString[event]
}

func (event StrapiWebhookEvents) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(toJson[event])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (event *StrapiWebhookEvents) UnmarshalJSON(bytes []byte) error {
	var str string
	err := json.Unmarshal(bytes, &str)

	if err != nil {
		return err
	}

	*event = toId[str]
	return nil
}
