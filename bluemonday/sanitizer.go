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
	p := bluemonday.UGCPolicy()

	// For youtube, instagram etc...
	p.AllowAttrs("url").OnElements("oembed") // ToDo: allow only some urls

	// Text display control - fonts, sizes, colors
	p.AllowAttrs("class", "style").OnElements("span", "ol", "ul", "figure") // ToDo: add matcher

	// Allow links opening in new tab
	p.AllowAttrs("target").
		Matching(regexp.MustCompile("_blank")).
		OnElements("a")

	p.AllowRelativeURLs(true)

	p.AllowStyling() // ToDo: use AllowStyles instead

	s.policy = p
}

func (s SanitizerService) Sanitize(html string) string {
	return s.policy.Sanitize(html)
}
