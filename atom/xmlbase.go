package atom

import (
	"fmt"
	"net/url"
)

// resolve u relative to b
func xmlBaseResolveUrl(b *url.URL, u string) (*url.URL, error) {
	relURL, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("gofeed/internal/shared: %w", err)
	} else if b == nil {
		return relURL, nil
	}

	if b.Path != "" && u != "" && b.Path[len(b.Path)-1] != '/' {
		// There's no reason someone would use a path in xml:base if they
		// didn't mean for it to be a directory
		b.Path += "/"
	}
	return b.ResolveReference(relURL), nil
}
