package bluemonday

import (
	"bytes"
	"fmt"
	"github.com/dyatlov/go-oembed/oembed"
	"github.com/kovansky/midas"
	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io/ioutil"
	"regexp"
	"strings"
)

var _ midas.SanitizerService = (*SanitizerService)(nil)

const iframePrefix = "https://oembed.link/"

var oembeds = oembed.NewOembed()

type SanitizerService struct {
	policy *bluemonday.Policy
}

func init() {
	providersFile, err := ioutil.ReadFile("providers_all.json")
	if err != nil {
		midas.ReportPanic(err)
	}

	err = oembeds.ParseProviders(bytes.NewReader(providersFile))
	if err != nil {
		midas.ReportPanic(err)
	}
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
	p.AllowAttrs("url").
		Matching(regexp.MustCompile("^https?:\\/\\/([-a-zA-Z0-9@:%_\\+~\\*#?&//=]*\\.)?((?:youtube\\.com|youtu.be))(\\/(?:[\\w\\-]+\\?v=|embed\\/|v\\/)?)([\\w\\-]+)(\\S+)?$")).
		OnElements("oembed") // ToDo: allow only some urls

	// ToDo: for oembeds: read oembeds and resolve them

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

	p.AllowStyles("width").
		Matching(bluemonday.NumberOrPercent).OnElements("figure")

	p.AllowAttrs("class").Matching(bluemonday.SpaceSeparatedTokens).OnElements(classAllowedElements...)

	// Allow links opening in new tab
	p.AllowAttrs("target").
		Matching(regexp.MustCompile("^_blank$")).
		OnElements("a")

	p.AllowRelativeURLs(true)
	p.RequireNoReferrerOnFullyQualifiedLinks(true)

	s.policy = p
}

func (s SanitizerService) Sanitize(htmlString string) string {
	return s.policy.Sanitize(htmlString)
}

func (s SanitizerService) Embed(htmlString string) (string, error) {
	dom, err := html.Parse(strings.NewReader(htmlString))
	if err != nil {
		return "", err
	}

	replaceEmbeds(dom)

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, dom)
	if err != nil {
		return "", err
	}

	return rendered.String(), nil
}

func replaceEmbeds(node *html.Node) {
	if node.Type == html.ElementNode && node.Data == "oembed" {
		iframeNode := generateIframe(node)

		node.Parent.InsertBefore(iframeNode, node)
		node.Parent.RemoveChild(node)
		return
	}

	// Search through all children
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		replaceEmbeds(child)
	}
}

func generateIframe(node *html.Node) *html.Node {
	var embedUrl string

	for _, attr := range node.Attr {
		if attr.Key == "url" {
			embedUrl = attr.Val
			break
		}
	}

	var provider *oembed.Item
	var oembedData = &oembed.Info{}

	provider = oembeds.FindItem(embedUrl)

	if provider != nil {
		if fetchedData, err := provider.FetchOembed(oembed.Options{URL: embedUrl}); err == nil {
			oembedData = fetchedData
		}
	}

	iframeNode := &html.Node{
		Type:     html.ElementNode,
		DataAtom: atom.Iframe,
		Data:     "iframe",
		Attr: []html.Attribute{
			{Key: "src", Val: fmt.Sprintf("%s%s", iframePrefix, embedUrl)},
			{Key: "sandbox", Val: "allow-scripts allow-same-origin"},
			{Key: "height", Val: fmt.Sprintf("%dpx", oembedData.ThumbnailHeight)},
			{Key: "width", Val: fmt.Sprintf("%dpx", oembedData.ThumbnailWidth)},
		},
	}

	return iframeNode
}
