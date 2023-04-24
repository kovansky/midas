/*
 * Copyright (c) 2023.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package concurrent

import (
	"github.com/kovansky/midas"
)

type List struct {
	processes map[string]*midas.Concurrent
}

func NewList() *List {
	return &List{make(map[string]*midas.Concurrent)}
}

// Add a new element to the list.
// If a process for the Site is already in the list, try to kill the process, remove it and add the new one
func (l *List) Add(concurrent midas.Concurrent) error {
	if l.Has(concurrent.Site().SiteName) == true {
		err := l.SafelyRemove(concurrent.Site().SiteName)
		if err != nil {
			return err
		}
	}
	l.processes[concurrent.Site().SiteName] = &concurrent
	return nil
}

// Has return true if the Site with the provided name has a running process.
func (l *List) Has(name string) bool {
	_, ok := l.processes[name]
	return ok
}

// SafelyRemove tries to kill the process for the provided Site and then removes it from the list
func (l *List) SafelyRemove(name string) error {
	process, err := l.Get(name)
	if err != nil {
		return err
	}
	(*process).Stop()
	l.Remove(name)
	return nil
}

// Remove the process for the given Site from the list without killing it (i.e. if we know it already ended).
func (l *List) Remove(name string) {
	delete(l.processes, name)
}

// Get the process for the Site with provided name.
func (l *List) Get(name string) (*midas.Concurrent, error) {
	if l.Has(name) {
		return l.processes[name], nil
	}

	return nil, midas.Errorf(midas.ErrProcessNotFound, "process for %s not found", name)
}
