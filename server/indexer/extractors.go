package indexer

import (
	"bytes"
	"errors"
	"io"
	"strings"

	"golang.org/x/net/html"
)

type Extractor interface {
	Match(*Document) bool
	Extract(*Document) error
}

var extractors []Extractor = []Extractor{
	&defaultExtractor{},
}

type defaultExtractor struct{}

func (e *defaultExtractor) Match(_ *Document) bool {
	return true
}

func (e *defaultExtractor) Extract(d *Document) error {
	r := bytes.NewReader([]byte(d.HTML))
	doc := html.NewTokenizer(r)
	inBody := false
	skip := false
	var text strings.Builder
	var currentTag string
out:
	for {
		tt := doc.Next()
		switch tt {
		case html.ErrorToken:
			err := doc.Err()
			if errors.Is(err, io.EOF) {
				break out
			}
			return errors.New("failed to parse html: " + err.Error())
		case html.SelfClosingTagToken, html.StartTagToken:
			tn, hasAttrs := doc.TagName()
			currentTag = string(tn)
			switch currentTag {
			case "body":
				inBody = true
			case "script", "style":
				skip = true
			case "link":
				var href string
				icon := false
				if !hasAttrs {
					break
				}
				for {
					aName, aVal, moreAttr := doc.TagAttr()
					if bytes.Equal(aName, []byte("href")) {
						href = string(aVal)
					}
					if bytes.Equal(aName, []byte("rel")) && bytes.Contains(aVal, []byte("icon")) {
						icon = true
					}
					if !moreAttr {
						break
					}
				}
				if icon && href != "" {
					d.faviconURL = fullURL(d.URL, href)
				}
			}
		case html.TextToken:
			if currentTag == "title" {
				d.Title += strings.TrimSpace(string(doc.Text()))
			}
			if inBody && !skip {
				text.Write(doc.Text())
			}
		case html.EndTagToken:
			tn, _ := doc.TagName()
			switch string(tn) {
			case "body":
				inBody = false
			case "script", "style":
				skip = false
			}
		}
	}
	d.Text = strings.TrimSpace(text.String())
	if d.Text == "" {
		return errors.New("no text found")
	}
	if d.Title == "" {
		return errors.New("no title found")
	}
	return nil
}
