package gofeed

import (
	"errors"
	"strconv"

	"github.com/dsh2dsh/gofeed/v2/atom"
	"github.com/dsh2dsh/gofeed/v2/internal/shared"
	"github.com/dsh2dsh/gofeed/v2/json"
	"github.com/dsh2dsh/gofeed/v2/options"
	"github.com/dsh2dsh/gofeed/v2/rss"
)

// Translator converts a particular feed (atom.Feed or rss.Feed of json.Feed)
// into the generic Feed struct
type Translator interface {
	Translate(feed any, opts *options.Parse) (*Feed, error)
}

// DefaultRSSTranslator converts an rss.Feed struct
// into the generic Feed struct.
//
// This default implementation defines a set of
// mapping rules between rss.Feed -> Feed
// for each of the fields in Feed.
type DefaultRSSTranslator struct{}

// Translate converts an RSS feed into the universal
// feed type.
func (t *DefaultRSSTranslator) Translate(feed any, opts *options.Parse) (*Feed, error) {
	rss, found := feed.(*rss.Feed)
	if !found {
		return nil, errors.New("Feed did not match expected type of *rss.Feed")
	}

	return &Feed{
		Title:           rss.GetTitle(),
		Description:     rss.GetDescription(),
		Link:            rss.GetLink(),
		Links:           rss.GetLinks(),
		FeedLink:        rss.GetFeedLink(),
		Updated:         rss.GetUpdated(),
		UpdatedParsed:   rss.GetUpdatedParsed(),
		Published:       rss.PubDate,
		PublishedParsed: rss.PubDateParsed,
		Author:          t.feedAuthor(rss),
		Authors:         t.feedAuthors(rss),
		Language:        rss.GetLanguage(),
		Image:           t.feedImage(rss),
		Copyright:       rss.GetCopyright(),
		Generator:       rss.Generator,
		Categories:      rss.GetCategories(),
		Items:           t.feedItems(rss),
		ITunesExt:       rss.ITunesExt,
		DublinCoreExt:   rss.DublinCoreExt,
		Extensions:      rss.Extensions,
		FeedVersion:     rss.Version,
		FeedType:        "rss",
	}, nil
}

func (t *DefaultRSSTranslator) translateFeedItem(rssItem *rss.Item) *Item {
	item := &Item{
		Title:           rssItem.GetTitle(),
		Description:     rssItem.GetDescription(),
		Content:         rssItem.Content,
		Links:           rssItem.Links,
		Published:       rssItem.GetPublished(),
		PublishedParsed: rssItem.GetPublishedParsed(),
		Author:          t.itemAuthor(rssItem),
		Authors:         t.itemAuthors(rssItem),
		GUID:            rssItem.GetGUID(),
		Image:           t.itemImage(rssItem),
		Categories:      rssItem.GetCategories(),
		Enclosures:      t.itemEnclosures(rssItem),
		DublinCoreExt:   rssItem.DublinCoreExt,
		ITunesExt:       rssItem.ITunesExt,
		Extensions:      rssItem.Extensions,
	}

	if len(item.Links) != 0 {
		item.Link = item.Links[0]
	}
	return item
}

func (t *DefaultRSSTranslator) feedAuthor(rss *rss.Feed) *Person {
	if name, address, ok := rss.GetAuthor(); ok {
		return &Person{
			Name:  name,
			Email: address,
		}
	}
	return nil
}

func (t *DefaultRSSTranslator) feedAuthors(rss *rss.Feed) []*Person {
	if author := t.feedAuthor(rss); author != nil {
		return []*Person{author}
	}
	return nil
}

func (t *DefaultRSSTranslator) feedImage(rss *rss.Feed) *Image {
	if img := rss.GetImage(); img != nil {
		return &Image{Title: img.Title, URL: img.URL}
	}
	return nil
}

func (t *DefaultRSSTranslator) feedItems(rss *rss.Feed) []*Item {
	items := make([]*Item, len(rss.Items))
	for i, item := range rss.Items {
		items[i] = t.translateFeedItem(item)
	}
	return items
}

func (t *DefaultRSSTranslator) itemAuthor(rssItem *rss.Item) *Person {
	if name, address, ok := rssItem.GetAuthor(); ok {
		return &Person{
			Name:  name,
			Email: address,
		}
	}
	return nil
}

func (t *DefaultRSSTranslator) itemAuthors(rssItem *rss.Item) []*Person {
	if author := t.itemAuthor(rssItem); author != nil {
		return []*Person{author}
	}
	return nil
}

func (t *DefaultRSSTranslator) itemImage(rssItem *rss.Item) *Image {
	if s := rssItem.ImageURL(); s != "" {
		return &Image{URL: s}
	}
	return nil
}

func (t *DefaultRSSTranslator) itemEnclosures(rssItem *rss.Item) []*Enclosure {
	if len(rssItem.Enclosures) == 0 {
		return nil
	}

	// Accumulate the enclosures
	enclosures := make([]*Enclosure, len(rssItem.Enclosures))
	for i, enc := range rssItem.Enclosures {
		enclosures[i] = &Enclosure{
			URL:    enc.URL,
			Type:   enc.Type,
			Length: enc.Length,
		}
	}
	return enclosures
}

// DefaultAtomTranslator converts an atom.Feed struct
// into the generic Feed struct.
//
// This default implementation defines a set of
// mapping rules between atom.Feed -> Feed
// for each of the fields in Feed.
type DefaultAtomTranslator struct{}

// Translate converts an Atom feed into the universal
// feed type.
func (t *DefaultAtomTranslator) Translate(feed any, opts *options.Parse) (*Feed, error) {
	atom, found := feed.(*atom.Feed)
	if !found {
		return nil, errors.New("Feed did not match expected type of *atom.Feed")
	}

	return &Feed{
		Title:         atom.Title,
		Description:   atom.Subtitle,
		Link:          atom.GetLink(),
		FeedLink:      atom.GetFeedLink(),
		Links:         atom.GetLinks(),
		Updated:       atom.Updated,
		UpdatedParsed: atom.UpdatedParsed,
		Author:        t.feedAuthor(atom),
		Authors:       t.feedAuthors(atom),
		Language:      atom.Language,
		Image:         t.feedImage(atom),
		Copyright:     atom.Rights,
		Categories:    atom.GetCategories(),
		Generator:     atom.GetGenerator(),
		Items:         t.feedItems(atom),
		Extensions:    atom.Extensions,
		FeedVersion:   atom.Version,
		FeedType:      "atom",
	}, nil
}

func (t *DefaultAtomTranslator) feedItem(entry *atom.Entry) *Item {
	return &Item{
		Title:           entry.Title,
		Description:     entry.Summary,
		Content:         entry.GetContent(),
		Link:            entry.GetLink(),
		Links:           entry.GetLinks(),
		Updated:         entry.Updated,
		UpdatedParsed:   entry.UpdatedParsed,
		Published:       entry.GetPublished(),
		PublishedParsed: entry.GetPublishedParsed(),
		Author:          t.itemAuthor(entry),
		Authors:         t.itemAuthors(entry),
		GUID:            entry.ID,
		Categories:      entry.GetCategories(),
		Enclosures:      t.itemEnclosures(entry),
		Extensions:      entry.Extensions,
	}
}

func (t *DefaultAtomTranslator) feedAuthor(atom *atom.Feed) *Person {
	if a := atom.GetAuthor(); a != nil {
		return &Person{Name: a.Name, Email: a.Email}
	}
	return nil
}

func (t *DefaultAtomTranslator) feedAuthors(atom *atom.Feed) []*Person {
	if len(atom.Authors) == 0 {
		return nil
	}

	authors := make([]*Person, len(atom.Authors))
	for i, a := range atom.Authors {
		authors[i] = &Person{
			Name:  a.Name,
			Email: a.Email,
		}
	}
	return authors
}

func (t *DefaultAtomTranslator) feedImage(atom *atom.Feed) *Image {
	if s := atom.ImageURL(); s != "" {
		return &Image{URL: s}
	}
	return nil
}

func (t *DefaultAtomTranslator) feedItems(atom *atom.Feed) []*Item {
	items := make([]*Item, len(atom.Entries))
	for i, entry := range atom.Entries {
		items[i] = t.feedItem(entry)
	}
	return items
}

func (t *DefaultAtomTranslator) itemAuthor(entry *atom.Entry) *Person {
	if a := entry.GetAuthor(); a != nil {
		return &Person{Name: a.Name, Email: a.Email}
	}
	return nil
}

func (t *DefaultAtomTranslator) itemAuthors(entry *atom.Entry) []*Person {
	if len(entry.Authors) == 0 {
		return nil
	}

	authors := make([]*Person, len(entry.Authors))
	for i, a := range entry.Authors {
		authors[i] = &Person{Name: a.Name, Email: a.Email}
	}
	return authors
}

func (t *DefaultAtomTranslator) itemEnclosures(entry *atom.Entry) []*Enclosure {
	if len(entry.Links) == 0 {
		return nil
	}

	var enclosures []*Enclosure //nolint:prealloc // not all links enclosures
	for _, e := range entry.Links {
		if e.Rel != "enclosure" {
			continue
		}
		enclosures = append(enclosures, &Enclosure{
			URL:    e.Href,
			Length: e.Length,
			Type:   e.Type,
		})
	}
	return enclosures
}

// DefaultJSONTranslator converts an json.Feed struct
// into the generic Feed struct.
//
// This default implementation defines a set of
// mapping rules between json.Feed -> Feed
// for each of the fields in Feed.
type DefaultJSONTranslator struct{}

// Translate converts an JSON feed into the universal
// feed type.
func (t *DefaultJSONTranslator) Translate(feed any, opts *options.Parse) (*Feed, error) {
	json, found := feed.(*json.Feed)
	if !found {
		return nil, errors.New("Feed did not match expected type of *json.Feed")
	}

	return &Feed{
		FeedVersion:     json.Version,
		Title:           json.Title,
		Link:            json.HomePageURL,
		FeedLink:        json.FeedURL,
		Links:           json.GetLinks(),
		Description:     json.Description,
		Image:           t.feedImage(json),
		Author:          t.feedAuthor(json),
		Authors:         t.feedAuthors(json),
		Language:        json.Language,
		Items:           t.feedItems(json),
		Updated:         json.GetUpdated(),
		UpdatedParsed:   json.GetUpdatedParsed(),
		Published:       json.GetPublished(),
		PublishedParsed: json.GetPublishedParsed(),
		FeedType:        "json",

		// TODO UserComment is missing in global Feed
		// TODO NextURL is missing in global Feed
		// TODO Favicon is missing in global Feed
		// TODO Exipred is missing in global Feed
		// TODO Hubs is not supported in json.Feed
		// TODO Extensions is not supported in json.Feed
	}, nil
}

func (t *DefaultJSONTranslator) feedItem(jsonItem *json.Item) *Item {
	return &Item{
		GUID:            jsonItem.ID,
		Link:            jsonItem.URL,
		Links:           jsonItem.Links(),
		Title:           jsonItem.Title,
		Content:         jsonItem.Content(),
		Description:     jsonItem.Summary,
		Image:           t.itemImage(jsonItem),
		Published:       jsonItem.DatePublished,
		PublishedParsed: jsonItem.PublishedParsed(),
		Updated:         jsonItem.DateModified,
		UpdatedParsed:   jsonItem.UpdatedParsed(),
		Author:          t.itemAuthor(jsonItem),
		Authors:         t.itemAuthors(jsonItem),
		Categories:      jsonItem.Tags,
		Enclosures:      t.itemEnclosures(jsonItem),

		// TODO ExternalURL is missing in global Feed
		// TODO BannerImage is missing in global Feed
	}
}

func (t *DefaultJSONTranslator) feedAuthor(json *json.Feed) *Person {
	if json.Author == nil {
		return nil
	}

	name, address := shared.ParseNameAddress(json.Author.Name)
	// Author.URL is missing in global feed
	// Author.Avatar is missing in global feed
	return &Person{Name: name, Email: address}
}

func (t *DefaultJSONTranslator) feedAuthors(json *json.Feed) []*Person {
	if json.Authors != nil {
		authors := make([]*Person, len(json.Authors))
		for i, a := range json.Authors {
			name, address := shared.ParseNameAddress(a.Name)
			authors[i] = &Person{Name: name, Email: address}
		}
		return authors
	}

	if author := t.feedAuthor(json); author != nil {
		return []*Person{author}
	}

	// Author.URL is missing in global feed
	// Author.Avatar is missing in global feed
	return nil
}

func (t *DefaultJSONTranslator) feedImage(json *json.Feed) *Image {
	// Using the Icon rather than the image
	//
	// icon (optional, string) is the URL of an image for the feed suitable to be
	// used in a timeline. It should be square and relatively large — such as 512
	// x 512
	if json.Icon != "" {
		return &Image{URL: json.Icon}
	}
	return nil
}

func (t *DefaultJSONTranslator) feedItems(json *json.Feed) []*Item {
	items := make([]*Item, len(json.Items))
	for i, it := range json.Items {
		items[i] = t.feedItem(it)
	}
	return items
}

func (t *DefaultJSONTranslator) itemAuthor(jsonItem *json.Item) *Person {
	if jsonItem.Author == nil {
		return nil
	}

	name, address := shared.ParseNameAddress(jsonItem.Author.Name)
	// Author.URL is missing in global feed
	// Author.Avatar is missing in global feed
	return &Person{Name: name, Email: address}
}

func (t *DefaultJSONTranslator) itemAuthors(jsonItem *json.Item) []*Person {
	if jsonItem.Authors != nil {
		authors := make([]*Person, len(jsonItem.Authors))
		for i, a := range jsonItem.Authors {
			name, address := shared.ParseNameAddress(a.Name)
			authors[i] = &Person{Name: name, Email: address}
		}
		return authors
	}

	if author := t.itemAuthor(jsonItem); author != nil {
		return []*Person{author}
	}
	// Author.URL is missing in global feed
	// Author.Avatar is missing in global feed
	return nil
}

func (t *DefaultJSONTranslator) itemImage(jsonItem *json.Item) *Image {
	if s := jsonItem.ImageURL(); s != "" {
		return &Image{URL: s}
	}
	return nil
}

func (t *DefaultJSONTranslator) itemEnclosures(jsonItem *json.Item) []*Enclosure {
	if jsonItem.Attachments == nil {
		return nil
	} else if len(*jsonItem.Attachments) == 0 {
		return nil
	}

	enclosures := make([]*Enclosure, len(*jsonItem.Attachments))
	for i, attachment := range *jsonItem.Attachments {
		// Title is not defined in global enclosure
		// SizeInBytes is not defined in global enclosure
		enclosures[i] = &Enclosure{
			URL:    attachment.URL,
			Type:   attachment.MimeType,
			Length: strconv.FormatInt(attachment.DurationInSeconds, 10),
		}
	}
	return enclosures
}
