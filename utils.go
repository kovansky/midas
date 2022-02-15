/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package midas

import "github.com/gosimple/slug"

// CreateSlug generates an url-safe string from title (or any other string) to be used as post/page slug.
func CreateSlug(title string) string {
	return slug.Make(title)
}
