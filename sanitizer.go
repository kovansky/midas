package midas

type SanitizerService interface {
	Sanitize(html string) string
	Embed(html string) (string, error)
}
