/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package midas

// Registry type is used to hold data from registries. It's structure is
// Id => Filename. So from this JSON:
//  {
//    "1": "sample-post.html"
//  }
// "1" would be a key and "sample-post.html" would be a value.
type Registry map[string]string

type RegistryService interface {
	OpenStorage() error
	CloseStorage()
	CreateStorage() error
	RemoveStorage() error
	Flush() error
	CreateEntry(id, filename string) error
	ReadEntry(id string) (string, error)
	UpdateEntry(id, newFilename string) error
	DeleteEntry(id string) error
}
