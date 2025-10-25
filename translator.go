package gofeed

import (
	"errors"
	"strconv"
	"time"

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
	return &Item{
		Title:           rssItem.GetTitle(),
		Description:     rssItem.GetDescription(),
		Content:         rssItem.Content,
		Link:            rssItem.Link,
		Links:           rssItem.GetLinks(),
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

	result := &Feed{}
	result.FeedVersion = json.Version
	result.Title = t.translateFeedTitle(json)
	result.Link = t.translateFeedLink(json)
	result.FeedLink = t.translateFeedFeedLink(json)
	result.Links = t.translateFeedLinks(json)
	result.Description = t.translateFeedDescription(json)
	result.Image = t.translateFeedImage(json)
	result.Author = t.translateFeedAuthor(json)
	result.Authors = t.translateFeedAuthors(json)
	result.Language = t.translateFeedLanguage(json)
	result.Items = t.translateFeedItems(json)
	result.Updated = t.translateFeedUpdated(json)
	result.UpdatedParsed = t.translateFeedUpdatedParsed(json)
	result.Published = t.translateFeedPublished(json)
	result.PublishedParsed = t.translateFeedPublishedParsed(json)
	result.FeedType = "json"
	// TODO UserComment is missing in global Feed
	// TODO NextURL is missing in global Feed
	// TODO Favicon is missing in global Feed
	// TODO Exipred is missing in global Feed
	// TODO Hubs is not supported in json.Feed
	// TODO Extensions is not supported in json.Feed
	return result, nil
}

func (t *DefaultJSONTranslator) translateFeedItem(jsonItem *json.Item) (item *Item) {
	item = &Item{}
	item.GUID = t.translateItemGUID(jsonItem)
	item.Link = t.translateItemLink(jsonItem)
	item.Links = t.translateItemLinks(jsonItem)
	item.Title = t.translateItemTitle(jsonItem)
	item.Content = t.translateItemContent(jsonItem)
	item.Description = t.translateItemDescription(jsonItem)
	item.Image = t.translateItemImage(jsonItem)
	item.Published = t.translateItemPublished(jsonItem)
	item.PublishedParsed = t.translateItemPublishedParsed(jsonItem)
	item.Updated = t.translateItemUpdated(jsonItem)
	item.UpdatedParsed = t.translateItemUpdatedParsed(jsonItem)
	item.Author = t.translateItemAuthor(jsonItem)
	item.Authors = t.translateItemAuthors(jsonItem)
	item.Categories = t.translateItemCategories(jsonItem)
	item.Enclosures = t.translateItemEnclosures(jsonItem)
	// TODO ExternalURL is missing in global Feed
	// TODO BannerImage is missing in global Feed
	return item
}

func (t *DefaultJSONTranslator) translateFeedTitle(json *json.Feed) (title string) {
	if json.Title != "" {
		title = json.Title
	}
	return title
}

func (t *DefaultJSONTranslator) translateFeedDescription(json *json.Feed) (desc string) {
	return json.Description
}

func (t *DefaultJSONTranslator) translateFeedLink(json *json.Feed) (link string) {
	if json.HomePageURL != "" {
		link = json.HomePageURL
	}
	return link
}

func (t *DefaultJSONTranslator) translateFeedFeedLink(json *json.Feed) (link string) {
	if json.FeedURL != "" {
		link = json.FeedURL
	}
	return link
}

func (t *DefaultJSONTranslator) translateFeedLinks(json *json.Feed) (links []string) {
	if json.HomePageURL != "" {
		links = append(links, json.HomePageURL)
	}
	if json.FeedURL != "" {
		links = append(links, json.FeedURL)
	}
	return links
}

func (t *DefaultJSONTranslator) translateFeedUpdated(json *json.Feed) (updated string) {
	if len(json.Items) > 0 {
		updated = json.Items[0].DateModified
	}
	return updated
}

func (t *DefaultJSONTranslator) translateFeedUpdatedParsed(json *json.Feed) (updated *time.Time) {
	if len(json.Items) > 0 {
		updateTime, err := shared.ParseDate(json.Items[0].DateModified)
		if err == nil {
			updated = &updateTime
		}
	}
	return updated
}

func (t *DefaultJSONTranslator) translateFeedPublished(json *json.Feed) (published string) {
	if len(json.Items) > 0 {
		published = json.Items[0].DatePublished
	}
	return published
}

func (t *DefaultJSONTranslator) translateFeedPublishedParsed(json *json.Feed) (published *time.Time) {
	if len(json.Items) > 0 {
		publishTime, err := shared.ParseDate(json.Items[0].DatePublished)
		if err == nil {
			published = &publishTime
		}
	}
	return published
}

func (t *DefaultJSONTranslator) translateFeedAuthor(json *json.Feed) (author *Person) {
	if json.Author != nil {
		name, address := shared.ParseNameAddress(json.Author.Name)
		author = &Person{}
		author.Name = name
		author.Email = address
	}
	// Author.URL is missing in global feed
	// Author.Avatar is missing in global feed
	return author
}

func (t *DefaultJSONTranslator) translateFeedAuthors(json *json.Feed) (authors []*Person) {
	if json.Authors != nil {
		authors = make([]*Person, 0, len(json.Authors))
		for _, a := range json.Authors {
			name, address := shared.ParseNameAddress(a.Name)
			author := &Person{}
			author.Name = name
			author.Email = address

			authors = append(authors, author)
		}
	} else if author := t.translateFeedAuthor(json); author != nil {
		authors = []*Person{author}
	}
	// Author.URL is missing in global feed
	// Author.Avatar is missing in global feed
	return authors
}

func (t *DefaultJSONTranslator) translateFeedLanguage(json *json.Feed) string {
	return json.Language
}

func (t *DefaultJSONTranslator) translateFeedImage(json *json.Feed) (image *Image) {
	// Using the Icon rather than the image
	// icon (optional, string) is the URL of an image for the feed suitable to be used in a timeline. It should be square and relatively large — such as 512 x 512
	if json.Icon != "" {
		image = &Image{}
		image.URL = json.Icon
	}
	return image
}

func (t *DefaultJSONTranslator) translateFeedItems(json *json.Feed) (items []*Item) {
	items = make([]*Item, 0, len(json.Items))
	for _, i := range json.Items {
		items = append(items, t.translateFeedItem(i))
	}
	return items
}

func (t *DefaultJSONTranslator) translateItemTitle(jsonItem *json.Item) (title string) {
	if jsonItem.Title != "" {
		title = jsonItem.Title
	}
	return title
}

func (t *DefaultJSONTranslator) translateItemDescription(jsonItem *json.Item) (desc string) {
	if jsonItem.Summary != "" {
		desc = jsonItem.Summary
	}
	return desc
}

func (t *DefaultJSONTranslator) translateItemContent(jsonItem *json.Item) (content string) {
	if jsonItem.ContentHTML != "" {
		content = jsonItem.ContentHTML
	} else if jsonItem.ContentText != "" {
		content = jsonItem.ContentText
	}
	return content
}

func (t *DefaultJSONTranslator) translateItemLink(jsonItem *json.Item) string {
	return jsonItem.URL
}

func (t *DefaultJSONTranslator) translateItemLinks(jsonItem *json.Item) (links []string) {
	if jsonItem.URL != "" {
		links = append(links, jsonItem.URL)
	}
	if jsonItem.ExternalURL != "" {
		links = append(links, jsonItem.ExternalURL)
	}
	return links
}

func (t *DefaultJSONTranslator) translateItemUpdated(jsonItem *json.Item) (updated string) {
	if jsonItem.DateModified != "" {
		updated = jsonItem.DateModified
	}
	return updated
}

func (t *DefaultJSONTranslator) translateItemUpdatedParsed(jsonItem *json.Item) (updated *time.Time) {
	if jsonItem.DateModified != "" {
		updatedTime, err := shared.ParseDate(jsonItem.DateModified)
		if err == nil {
			updated = &updatedTime
		}
	}
	return updated
}

func (t *DefaultJSONTranslator) translateItemPublished(jsonItem *json.Item) (pubDate string) {
	if jsonItem.DatePublished != "" {
		pubDate = jsonItem.DatePublished
	}
	return pubDate
}

func (t *DefaultJSONTranslator) translateItemPublishedParsed(jsonItem *json.Item) (pubDate *time.Time) {
	if jsonItem.DatePublished != "" {
		publishTime, err := shared.ParseDate(jsonItem.DatePublished)
		if err == nil {
			pubDate = &publishTime
		}
	}
	return pubDate
}

func (t *DefaultJSONTranslator) translateItemAuthor(jsonItem *json.Item) (author *Person) {
	if jsonItem.Author != nil {
		name, address := shared.ParseNameAddress(jsonItem.Author.Name)
		author = &Person{}
		author.Name = name
		author.Email = address
	}
	// Author.URL is missing in global feed
	// Author.Avatar is missing in global feed
	return author
}

func (t *DefaultJSONTranslator) translateItemAuthors(jsonItem *json.Item) (authors []*Person) {
	if jsonItem.Authors != nil {
		authors = make([]*Person, 0, len(jsonItem.Authors))
		for _, a := range jsonItem.Authors {
			name, address := shared.ParseNameAddress(a.Name)
			author := &Person{}
			author.Name = name
			author.Email = address

			authors = append(authors, author)
		}
	} else if author := t.translateItemAuthor(jsonItem); author != nil {
		authors = []*Person{author}
	}
	// Author.URL is missing in global feed
	// Author.Avatar is missing in global feed
	return authors
}

func (t *DefaultJSONTranslator) translateItemGUID(jsonItem *json.Item) (guid string) {
	if jsonItem.ID != "" {
		guid = jsonItem.ID
	}
	return guid
}

func (t *DefaultJSONTranslator) translateItemImage(jsonItem *json.Item) (image *Image) {
	if jsonItem.Image != "" {
		image = &Image{}
		image.URL = jsonItem.Image
	} else if jsonItem.BannerImage != "" {
		image = &Image{}
		image.URL = jsonItem.BannerImage
	}
	return image
}

func (t *DefaultJSONTranslator) translateItemCategories(jsonItem *json.Item) (categories []string) {
	if len(jsonItem.Tags) > 0 {
		categories = jsonItem.Tags
	}
	return categories
}

func (t *DefaultJSONTranslator) translateItemEnclosures(jsonItem *json.Item) (enclosures []*Enclosure) {
	if jsonItem.Attachments != nil {
		for _, attachment := range *jsonItem.Attachments {
			e := &Enclosure{}
			e.URL = attachment.URL
			e.Type = attachment.MimeType
			e.Length = strconv.FormatInt(attachment.DurationInSeconds, 10)
			// Title is not defined in global enclosure
			// SizeInBytes is not defined in global enclosure
			enclosures = append(enclosures, e)
		}
	}
	return enclosures
}
