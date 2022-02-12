package bluemonday

import (
	"github.com/kovansky/midas"
	"github.com/microcosm-cc/bluemonday"
	"regexp"
)

var _ midas.SanitizerService = (*SanitizerService)(nil)

type SanitizerService struct {
	policy *bluemonday.Policy
}

func NewSanitizerService() *SanitizerService {
	sanitizerService := &SanitizerService{}

	sanitizerService.injectDefaultPolicy()

	return sanitizerService
}

func (s *SanitizerService) injectDefaultPolicy() {
	textElements := []string{
		"span", "p",
	}

	classAllowedElements := append(textElements,
		"mark",
		"ol", "ul",
		"figure")

	p := bluemonday.UGCPolicy()

	// For youtube, instagram etc...
	p.AllowAttrs("url").OnElements("oembed") // ToDo: allow only some urls

	// Text display control
	p.AllowStyles("font-family").
		Matching(regexp.MustCompile("(?i)^[a-z0-9\\-_ ,'\\\"]+$")).
		OnElements(textElements...)

	p.AllowStyles("color"). // Hex (3 chars, 6 chars), rgb, rgba (number or percentage), hsl are supported
				Matching(regexp.MustCompile("^(#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3}))|(rgb\\( *\\d{1,3}%? *, *\\d{1,3}%? *, *\\d{1,3}%? *\\))|(rgba\\( *\\d{1,3}%? *, *\\d{1,3}%? *, *\\d{1,3}%? *, *((1(\\.0+)?)|(0\\.\\d+)) *\\))|(hsl\\( *\\d{1,3} *, *\\d{1,3}% *, *\\d{1,3}% *\\))$")).
				OnElements(textElements...)

	p.AllowStyles("list-style-types").
		MatchingEnum("disc", "circle", "square", "decimal", "decimal-leading-zero",
			"lower-roman", "upper-roman", "lower-greek", "lower-latin", "upper-latin", "armenian",
			"georgian", "lower-alpha", "upper-alpha", "none").
		OnElements("ul", "ol")

	p.AllowAttrs("class").Matching(bluemonday.SpaceSeparatedTokens).OnElements(classAllowedElements...)

	// Allow links opening in new tab
	p.AllowAttrs("target").
		Matching(regexp.MustCompile("^_blank$")).
		OnElements("a")

	p.AllowRelativeURLs(true)
	p.RequireNoReferrerOnFullyQualifiedLinks(true)

	s.policy = p
}

func (s SanitizerService) Sanitize(html string) string {
	return s.policy.Sanitize(html)
}
