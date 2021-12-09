package models

import "github.com/gosimple/slug"

func createSlug(title string) string {
	return slug.Make(title)
}
