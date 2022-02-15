/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package midas

type Payload interface {
	Event() string
	Metadata() map[string]interface{}
	Entry() map[string]interface{}
	Raw() interface{}
}
