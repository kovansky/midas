/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package midas

import "encoding/json"

type Payload interface {
	json.Unmarshaler
	json.Marshaler
	Event() string
	Metadata() map[string]interface{}
	Entry() map[string]interface{}
	Raw() interface{}
}
