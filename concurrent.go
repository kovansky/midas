/*
 * Copyright (c) 2023.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package midas

type ConcurrentList interface {
	Add(concurrent Concurrent) error
	Has(name string) bool
	SafelyRemove(name string) error
	Remove(name string)
	Get(name string) (*Concurrent, error)
}

type Concurrent interface {
	Stop()
	Site() Site
}
