package midas

type SanitizerService interface {
	Sanitize(html string) string
}
