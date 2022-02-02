package strapi

import (
	"bytes"
	"encoding/json"
)

type Event int64

const (
	Undefined Event = iota
	Create
	Update
	Delete
	Publish
	Unpublish
)

var toJson = map[Event]string{
	Undefined: "",
	Create:    "entry.create",
	Update:    "entry.update",
	Delete:    "entry.delete",
	Publish:   "entry.publish",
	Unpublish: "entry.unpublish",
}

var toString = map[Event]string{
	Undefined: "",
	Create:    "Create",
	Update:    "Update",
	Delete:    "Delete",
	Publish:   "Publish",
	Unpublish: "Unpublish",
}

var toId = map[string]Event{
	"entry.create":    Create,
	"entry.update":    Update,
	"entry.delete":    Delete,
	"entry.publish":   Publish,
	"entry.unpublish": Unpublish,
}

func (event Event) String() string {
	return toString[event]
}

func (event Event) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(toJson[event])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (event *Event) UnmarshalJSON(bytes []byte) error {
	var str string
	err := json.Unmarshal(bytes, &str)

	if err != nil {
		return err
	}

	var ok bool

	if *event, ok = toId[str]; !ok {
		*event = Undefined
	}
	return nil
}
