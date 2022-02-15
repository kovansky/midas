/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package bluemonday

import (
	"testing"
)

func TestSanitizerService_Sanitize(t *testing.T) {
	sanitizer := NewSanitizerService()

	type args struct {
		html string
	}
	tests := map[string][]struct {
		name string
		args args
		want string
	}{
		"Elements": {
			{"Allowed", args{`<p></p>`}, `<p></p>`},
			{"Disallowed", args{`<iframe src="https://google.com"></iframe>`}, ``},
		},
		"Styles": {
			{
				"fontOnAllowed",
				args{`<p style="font-family: 'Courier New', Courier, monospace"></p>`},
				`<p style="font-family: &#39;Courier New&#39;, Courier, monospace"></p>`,
			},
			{
				"fontOnDisallowed",
				args{`<div style="font-family: 'Courier New', Courier, monospace"></div>`},
				`<div></div>`,
			},
			{
				"fontNotMatching",
				args{`<p style="font-family: 'Courier New', url(harmful/url)"></p>`},
				`<p></p>`,
			},
			{
				"colorHsl",
				args{`<span style="color: hsl(90,75%,60%);"></span>`},
				`<span style="color: hsl(90,75%,60%)"></span>`,
			},
			{
				"colorRgbNumeric",
				args{`<span style="color: rgb(100, 100, 100);"></span>`},
				`<span style="color: rgb(100, 100, 100)"></span>`,
			},
			{
				"colorRgbPercentage",
				args{`<span style="color: rgb(100%,100%,100%);"></span>`},
				`<span style="color: rgb(100%,100%,100%)"></span>`,
			},
			{
				"colorRgbaNumeric",
				args{`<span style="color: rgba(100,100,100,0.5);"></span>`},
				`<span style="color: rgba(100,100,100,0.5)"></span>`,
			},
			{
				"colorRgbaPercentage",
				args{`<span style="color: rgba(100%,100%,100%,0.1);"></span>`},
				`<span style="color: rgba(100%,100%,100%,0.1)"></span>`,
			},
			{
				"colorHex",
				args{`<span style="color: #FF90cc;"></span>`},
				`<span style="color: #FF90cc"></span>`,
			},
			{
				"colorShortHex",
				args{`<span style="color: #F9c;"></span>`},
				`<span style="color: #F9c"></span>`,
			},
			{
				"colorOnDisallowed",
				args{`<span style="color: hsl(90,75%,60%);"></span>`},
				`<span style="color: hsl(90,75%,60%)"></span>`,
			},
			{
				"colorHslNotMatching",
				args{`<span style="color: hsl(100%, 14, 156);"></span>`},
				`<span></span>`,
			},
			{
				"colorRgbaNotMatching",
				args{`<span style="color: rgba(100%,100%,100%,80%);"></span>`},
				`<span></span>`,
			},
			{
				"colorHexNotMatching",
				args{`<span style="color: #xyz;"></span>`},
				`<span></span>`,
			},
			{
				"colorNotMatching",
				args{`<span style="color: url(harmful/url);"></span>`},
				`<span></span>`,
			},
			{
				"colorOnDisallowed",
				args{`<div style="color: #ffffff;"></div>`},
				`<div></div>`,
			},
			{
				"listTypesOnAllowed",
				args{`<ul style="list-style-type:circle;"><li></li></ul>`},
				`<ul style="list-style-type: circle"><li></li></ul>`,
			},
			{
				"listTypesOnDisallowed",
				args{`<p style="list-style-type:circle;"></p>`},
				`<p></p>`,
			},
			{
				"listTypesNotMatching",
				args{`<ul style="list-style-type:diagonal-gradient;"><li></li></ul>`},
				`<ul><li></li></ul>`,
			},
			{
				"widthPx",
				args{`<figure style="width: 50px"></figure>`},
				`<figure style="width: 50px"></figure>`,
			},
			{
				"widthPercentage",
				args{`<figure style="width: 50%"></figure>`},
				`<figure style="width: 50%"></figure>`,
			},
			{
				"widthAuto",
				args{`<figure style="width: auto"></figure>`},
				`<figure style="width: auto"></figure>`,
			},
			{
				"widthNotMatching",
				args{`<figure style="width: url(harmful/url)"></figure>`},
				`<figure></figure>`,
			},
			{
				"widthOnDisallowed",
				args{`<span style="width: url(harmful/url)"></span>`},
				`<span></span>`,
			},
		},
		"Classes": {
			{
				"OnAllowed",
				args{`<mark class="mark-pink"></mark>`},
				`<mark class="mark-pink"></mark>`,
			},
			{
				"MultipleOnAllowed",
				args{`<p class="bg-primary text-big"></p>`},
				`<p class="bg-primary text-big"></p>`,
			},
			{
				"NotMatching",
				args{`<mark class="url(harmful/url)"></mark>`},
				`<mark></mark>`,
			},
			{
				"OnDisallowed",
				args{`<div class="text-big"></div>`},
				`<div></div>`,
			},
		},
		"Anchors": {
			{
				"AbsoluteUrl",
				args{`<a href="https://google.com">Test</a>`},
				`<a href="https://google.com" rel="nofollow noreferrer">Test</a>`,
			},
			{
				"AbsoluteUrlIncorrectProtocol",
				args{`<a href="ftp://google.com">Test</a>`},
				`Test`,
			},
			{
				"AbsoluteUrlCustomRels",
				args{`<a href="https://google.com" rel="author">Test</a>`},
				`<a href="https://google.com" rel="author nofollow noreferrer">Test</a>`,
			},
			{
				"RelativeUrl",
				args{`<a href="/index">Test</a>`},
				`<a href="/index" rel="nofollow">Test</a>`,
			},
			{
				"TargetBlank",
				args{`<a href="https://google.com" target="_blank">Test</a>`},
				`<a href="https://google.com" target="_blank" rel="nofollow noreferrer noopener">Test</a>`,
			},
			{
				"TargetForbidden",
				args{`<a href="/index" target="_top">Test</a>`},
				`<a href="/index" rel="nofollow">Test</a>`,
			},
			{
				"NoHref",
				args{`<a>Test</a>`},
				`Test`,
			},
			{
				"OnlyNamed",
				args{`<a name="test">Test</a>`},
				`Test`,
			},
		},
	}
	for tname, table := range tests {
		t.Run(tname, func(t *testing.T) {
			for _, tt := range table {
				t.Run(tt.name, func(t *testing.T) {
					if got := sanitizer.Sanitize(tt.args.html); got != tt.want {
						t.Errorf("Sanitize() = %v, want %v", got, tt.want)
					}
				})
			}
		})
	}
}
