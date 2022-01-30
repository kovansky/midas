package strapi2hugo

import "github.com/gosimple/slug"

func CreateSlug(title string) string {
	return slug.Make(title)
}

func Contains(haystack []string, needle string) bool {
	for _, value := range haystack {
		if value == needle {
			return true
		}
	}

	return false
}
