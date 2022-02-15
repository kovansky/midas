/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package midas

type Config struct {
	Domain       string          `json:"domain"`
	Addr         string          `json:"addr"`
	RollbarToken string          `json:"rollbarToken"`
	Sites        map[string]Site `json:"sites"` // [api key] => site
}
