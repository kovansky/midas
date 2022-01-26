package strapi

import (
	"bytes"
	"encoding/json"
)

type event int64

const (
	Undefined event = iota
	Create
	Update
	Delete
	Publish
	Unpublish
)

var toJson = map[event]string{
	Undefined: "",
	Create:    "entry.create",
	Update:    "entry.update",
	Delete:    "entry.delete",
	Publish:   "entry.publish",
	Unpublish: "entry.unpublish",
}

var toString = map[event]string{
	Undefined: "",
	Create:    "Create",
	Update:    "Update",
	Delete:    "Delete",
	Publish:   "Publish",
	Unpublish: "Unpublish",
}

var toId = map[string]event{
	"entry.create":    Create,
	"entry.update":    Update,
	"entry.delete":    Delete,
	"entry.publish":   Publish,
	"entry.unpublish": Unpublish,
}

func (event event) String() string {
	return toString[event]
}

func (event event) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(toJson[event])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (event *event) UnmarshalJSON(bytes []byte) error {
	var str string
	err := json.Unmarshal(bytes, &str)

	if err != nil {
		return err
	}

	*event = toId[str]
	return nil
}
