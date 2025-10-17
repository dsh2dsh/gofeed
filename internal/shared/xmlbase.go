package shared

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	xpp "github.com/mmcdole/goxpp"
)

// List of xml attributes that contain URIs to be resolved relative to
// xml:base
// From the Atom spec https://tools.ietf.org/html/rfc4287
var uriAttrs = map[string]bool{
	"href":   true,
	"scheme": true,
	"src":    true,
	"uri":    true,
}

// XMLBase.NextTag iterates through the tokens until it reaches a StartTag or
// EndTag. It resolves urls in tag attributes relative to the current xml:base.
//
// NextTag is similar to goxpp's NextTag method except it wont throw an error
// if the next immediate token isnt a Start/EndTag.  Instead, it will continue
// to consume tokens until it hits a Start/EndTag or EndDocument.
func NextTag(p *xpp.XMLPullParser) (event xpp.XMLEventType, err error) {
	for {
		event, err = p.Next()
		if err != nil {
			return event, fmt.Errorf("gofeed/internal/shared: %w", err)
		}

		if event == xpp.EndTag {
			break
		}

		if event == xpp.StartTag {
			err = resolveAttrs(p)
			if err != nil {
				return event, err
			}

			break
		}

		if event == xpp.EndDocument {
			return event, errors.New("failed to find NextTag before reaching the end of the document")
		}

	}
	return event, nil
}

// resolve relative URI attributes according to xml:base
func resolveAttrs(p *xpp.XMLPullParser) error {
	for i, attr := range p.Attrs {
		lowerName := strings.ToLower(attr.Name.Local)
		if uriAttrs[lowerName] {
			absURL, err := XmlBaseResolveUrl(p.BaseStack.Top(), attr.Value)
			if err == nil && absURL != nil {
				p.Attrs[i].Value = absURL.String()
			}
			// Continue processing even if URL resolution fails (e.g., for non-HTTP URIs like at://)
		}
	}
	return nil
}

// resolve u relative to b
func XmlBaseResolveUrl(b *url.URL, u string) (*url.URL, error) {
	relURL, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("gofeed/internal/shared: %w", err)
	}

	if b == nil {
		return relURL, nil
	}

	if b.Path != "" && u != "" && b.Path[len(b.Path)-1] != '/' {
		// There's no reason someone would use a path in xml:base if they
		// didn't mean for it to be a directory
		b.Path += "/"
	}
	absURL := b.ResolveReference(relURL)
	return absURL, nil
}
