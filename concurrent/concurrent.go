/*
 * Copyright (c) 2023.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package concurrent

import (
	"context"
	"github.com/kovansky/midas"
)

type Concurrent struct {
	site   midas.Site
	cancel context.CancelFunc
}

func New(site midas.Site, cancel context.CancelFunc) *Concurrent {
	return &Concurrent{
		site:   site,
		cancel: cancel,
	}
}

func (c *Concurrent) Stop() {
	c.cancel()
}

func (c *Concurrent) Site() midas.Site {
	return c.site
}
