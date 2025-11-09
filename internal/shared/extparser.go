package shared

import (
	"fmt"
	"strings"

	xpp "github.com/dsh2dsh/goxpp/v2"

	"github.com/dsh2dsh/gofeed/v2/ext"
)

var (
	emptyAttrs    = map[string]string{}
	emptyChildren = map[string][]ext.Extension{}
)

// ParseExtension parses the current element of the
// XMLPullParser as an extension element and updates
// the extension map
func ParseExtension(fe ext.Extensions, p *xpp.XMLPullParser) (ext.Extensions, error) {
	prefix := PrefixForNamespace(p.Space, p)

	result, err := parseExtensionElement(p)
	if err != nil {
		return nil, err
	}

	if fe == nil {
		fe = make(ext.Extensions, 1)
	}

	// Ensure the extension prefix map exists
	if m, ok := fe[prefix]; ok {
		m[p.Name] = append(m[p.Name], result)
	} else {
		fe[prefix] = map[string][]ext.Extension{p.Name: {result}}
	}
	return fe, nil
}

func parseExtensionElement(p *xpp.XMLPullParser) (e ext.Extension, err error) {
	if err = p.Expect(xpp.StartTag, "*"); err != nil {
		return e, fmt.Errorf("gofeed/internal/shared: %w", err)
	}

	e.Name = p.Name
	e.Attrs = emptyAttrs
	e.Children = emptyChildren

	if n := len(p.Attrs); n != 0 {
		e.Attrs = make(map[string]string, n)
		for _, attr := range p.Attrs {
			// TODO: Alright that we are stripping
			// namespace information from attributes ?
			e.Attrs[attr.Name.Local] = attr.Value
		}
	}

	var text1 string
	var text2 strings.Builder

	for {
		tok, err := p.Next()
		if err != nil {
			return e, fmt.Errorf("gofeed/internal/shared: %w", err)
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {
			child, err := parseExtensionElement(p)
			if err != nil {
				return e, err
			}
			if len(e.Children) == 0 {
				e.Children = map[string][]ext.Extension{child.Name: {child}}
			} else {
				e.Children[child.Name] = append(e.Children[child.Name], child)
			}
			continue
		} else if tok != xpp.Text {
			continue
		}

		switch {
		case text1 == "":
			text1 = p.Text()
		case text2.Len() == 0:
			text2.WriteString(text1)
			fallthrough
		default:
			text2.WriteString(p.Text())
		}
	}

	if text2.Len() == 0 {
		e.Value = strings.TrimSpace(text1)
	} else {
		e.Value = strings.TrimSpace(text2.String())
	}

	if err = p.Expect(xpp.EndTag, e.Name); err != nil {
		return e, fmt.Errorf("gofeed/internal/shared: %w", err)
	}
	return e, nil
}

func PrefixForNamespace(space string, p *xpp.XMLPullParser) string {
	// First we check if the global namespace map
	// contains an entry for this namespace/prefix.
	// This way we can use the canonical prefix for this
	// ns instead of the one defined in the feed.
	if prefix, ok := canonicalNamespaces[space]; ok {
		return prefix
	}

	// Next we check if the feed itself defined this
	// this namespace and return it if we have a result.
	if prefix, ok := p.Spaces[space]; ok {
		return prefix
	}

	// Lastly, any namespace which is not defined in the
	// the feed will be the prefix itself when using Go's
	// xml.Decoder.Token() method.
	return space
}

// Namespaces taken from github.com/kurtmckee/feedparser
// These are used for determining canonical name space prefixes
// for many of the popular RSS/Atom extensions.
//
// These canonical prefixes override any prefixes used in the feed itself.
var canonicalNamespaces = map[string]string{
	"http://webns.net/mvcb/":                                         "admin",
	"http://purl.org/rss/1.0/modules/aggregation/":                   "ag",
	"http://purl.org/rss/1.0/modules/annotate/":                      "annotate",
	"http://media.tangent.org/rss/1.0/":                              "audio",
	"http://backend.userland.com/blogChannelModule":                  "blogChannel",
	"http://creativecommons.org/ns#license":                          "cc",
	"http://web.resource.org/cc/":                                    "cc",
	"http://cyber.law.harvard.edu/rss/creativeCommonsRssModule.html": "creativeCommons",
	"http://backend.userland.com/creativeCommonsRssModule":           "creativeCommons",
	"http://purl.org/rss/1.0/modules/company":                        "co",
	"http://purl.org/rss/1.0/modules/content/":                       "content",
	"http://my.theinfo.org/changed/1.0/rss/":                         "cp",
	"http://purl.org/dc/elements/1.1/":                               "dc",
	"http://purl.org/dc/terms/":                                      "dcterms",
	"http://purl.org/rss/1.0/modules/email/":                         "email",
	"http://purl.org/rss/1.0/modules/event/":                         "ev",
	"http://rssnamespace.org/feedburner/ext/1.0":                     "feedburner",
	"http://freshmeat.net/rss/fm/":                                   "fm",
	"http://xmlns.com/foaf/0.1/":                                     "foaf",
	"http://www.w3.org/2003/01/geo/wgs84_pos#":                       "geo",
	"http://www.georss.org/georss":                                   "georss",
	"http://www.opengis.net/gml":                                     "gml",
	"http://postneo.com/icbm/":                                       "icbm",
	"http://purl.org/rss/1.0/modules/image/":                         "image",
	"http://www.itunes.com/DTDs/PodCast-1.0.dtd":                     "itunes",
	"http://example.com/DTDs/PodCast-1.0.dtd":                        "itunes",
	"http://purl.org/rss/1.0/modules/link/":                          "l",
	"http://search.yahoo.com/mrss":                                   "media",
	"http://search.yahoo.com/mrss/":                                  "media",
	"http://madskills.com/public/xml/rss/module/pingback/":           "pingback",
	"http://prismstandard.org/namespaces/1.2/basic/":                 "prism",
	"http://www.w3.org/1999/02/22-rdf-syntax-ns#":                    "rdf",
	"http://www.w3.org/2000/01/rdf-schema#":                          "rdfs",
	"http://purl.org/rss/1.0/modules/reference/":                     "ref",
	"http://purl.org/rss/1.0/modules/richequiv/":                     "reqv",
	"http://purl.org/rss/1.0/modules/search/":                        "search",
	"http://purl.org/rss/1.0/modules/slash/":                         "slash",
	"http://schemas.xmlsoap.org/soap/envelope/":                      "soap",
	"http://purl.org/rss/1.0/modules/servicestatus/":                 "ss",
	"http://hacks.benhammersley.com/rss/streaming/":                  "str",
	"http://purl.org/rss/1.0/modules/subscription/":                  "sub",
	"http://purl.org/rss/1.0/modules/syndication/":                   "sy",
	"http://schemas.pocketsoap.com/rss/myDescModule/":                "szf",
	"http://purl.org/rss/1.0/modules/taxonomy/":                      "taxo",
	"http://purl.org/rss/1.0/modules/threading/":                     "thr",
	"http://purl.org/rss/1.0/modules/textinput/":                     "ti",
	"http://madskills.com/public/xml/rss/module/trackback/":          "trackback",
	"http://wellformedweb.org/commentAPI/":                           "wfw",
	"http://purl.org/rss/1.0/modules/wiki/":                          "wiki",
	"http://www.w3.org/1999/xhtml":                                   "xhtml",
	"http://www.w3.org/1999/xlink":                                   "xlink",
	"http://www.w3.org/XML/1998/namespace":                           "xml",
	"http://podlove.org/simple-chapters":                             "psc",
}
