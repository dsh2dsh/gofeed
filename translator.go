package gofeed

import (
	"errors"
	"strconv"
	"strings"
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

	result := &Feed{}
	result.Title = t.translateFeedTitle(atom)
	result.Description = t.translateFeedDescription(atom)
	result.Link = t.translateFeedLink(atom)
	result.FeedLink = t.translateFeedFeedLink(atom)
	result.Links = t.translateFeedLinks(atom)
	result.Updated = t.translateFeedUpdated(atom)
	result.UpdatedParsed = t.translateFeedUpdatedParsed(atom)
	result.Author = t.translateFeedAuthor(atom)
	result.Authors = t.translateFeedAuthors(atom)
	result.Language = t.translateFeedLanguage(atom)
	result.Image = t.translateFeedImage(atom)
	result.Copyright = t.translateFeedCopyright(atom)
	result.Categories = t.translateFeedCategories(atom)
	result.Generator = t.translateFeedGenerator(atom)
	result.Items = t.translateFeedItems(atom)
	result.Extensions = atom.Extensions
	result.FeedVersion = atom.Version
	result.FeedType = "atom"
	return result, nil
}

func (t *DefaultAtomTranslator) translateFeedItem(entry *atom.Entry) (item *Item) {
	item = &Item{}
	item.Title = t.translateItemTitle(entry)
	item.Description = t.translateItemDescription(entry)
	item.Content = t.translateItemContent(entry)
	item.Link = t.translateItemLink(entry)
	item.Links = t.translateItemLinks(entry)
	item.Updated = t.translateItemUpdated(entry)
	item.UpdatedParsed = t.translateItemUpdatedParsed(entry)
	item.Published = t.translateItemPublished(entry)
	item.PublishedParsed = t.translateItemPublishedParsed(entry)
	item.Author = t.translateItemAuthor(entry)
	item.Authors = t.translateItemAuthors(entry)
	item.GUID = t.translateItemGUID(entry)
	item.Image = t.translateItemImage(entry)
	item.Categories = t.translateItemCategories(entry)
	item.Enclosures = t.translateItemEnclosures(entry)
	item.Extensions = entry.Extensions
	return item
}

func (t *DefaultAtomTranslator) translateFeedTitle(atom *atom.Feed) (title string) {
	return atom.Title
}

func (t *DefaultAtomTranslator) translateFeedDescription(atom *atom.Feed) (desc string) {
	return atom.Subtitle
}

func (t *DefaultAtomTranslator) translateFeedLink(atom *atom.Feed) (link string) {
	l := t.firstLinkWithType("alternate", atom.Links)
	if l != nil {
		link = l.Href
	}
	return link
}

func (t *DefaultAtomTranslator) translateFeedFeedLink(atom *atom.Feed) (link string) {
	feedLink := t.firstLinkWithType("self", atom.Links)
	if feedLink != nil {
		link = feedLink.Href
	}
	return link
}

func (t *DefaultAtomTranslator) translateFeedLinks(atom *atom.Feed) (links []string) {
	for _, l := range atom.Links {
		if l.Rel == "" || l.Rel == "alternate" || l.Rel == "self" {
			links = append(links, l.Href)
		}
	}
	return links
}

func (t *DefaultAtomTranslator) translateFeedUpdated(atom *atom.Feed) (updated string) {
	return atom.Updated
}

func (t *DefaultAtomTranslator) translateFeedUpdatedParsed(atom *atom.Feed) (updated *time.Time) {
	return atom.UpdatedParsed
}

func (t *DefaultAtomTranslator) translateFeedAuthor(atom *atom.Feed) (author *Person) {
	a := t.firstPerson(atom.Authors)
	if a != nil {
		feedAuthor := Person{}
		feedAuthor.Name = a.Name
		feedAuthor.Email = a.Email
		author = &feedAuthor
	}
	return author
}

func (t *DefaultAtomTranslator) translateFeedAuthors(atom *atom.Feed) (authors []*Person) {
	if atom.Authors != nil {
		authors = make([]*Person, 0, len(atom.Authors))

		for _, a := range atom.Authors {
			authors = append(authors, &Person{
				Name:  a.Name,
				Email: a.Email,
			})
		}
	}

	return authors
}

func (t *DefaultAtomTranslator) translateFeedLanguage(atom *atom.Feed) (language string) {
	return atom.Language
}

func (t *DefaultAtomTranslator) translateFeedImage(atom *atom.Feed) (image *Image) {
	if atom.Logo != "" {
		feedImage := Image{}
		feedImage.URL = atom.Logo
		image = &feedImage
	} else if atom.Icon != "" {
		feedImage := Image{}
		feedImage.URL = atom.Icon
		image = &feedImage
	}
	return image
}

func (t *DefaultAtomTranslator) translateFeedCopyright(atom *atom.Feed) (rights string) {
	return atom.Rights
}

func (t *DefaultAtomTranslator) translateFeedGenerator(atom *atom.Feed) (generator string) {
	if atom.Generator != nil {
		if atom.Generator.Value != "" {
			generator += atom.Generator.Value
		}
		if atom.Generator.Version != "" {
			generator += " v" + atom.Generator.Version
		}
		if atom.Generator.URI != "" {
			generator += " " + atom.Generator.URI
		}
		generator = strings.TrimSpace(generator)
	}
	return generator
}

func (t *DefaultAtomTranslator) translateFeedCategories(atom *atom.Feed) (categories []string) {
	if atom.Categories != nil {
		categories = make([]string, 0, len(atom.Categories))
		for _, c := range atom.Categories {
			if c.Label != "" {
				categories = append(categories, c.Label)
			} else {
				categories = append(categories, c.Term)
			}
		}
	}
	return categories
}

func (t *DefaultAtomTranslator) translateFeedItems(atom *atom.Feed) (items []*Item) {
	items = make([]*Item, 0, len(atom.Entries))
	for _, entry := range atom.Entries {
		items = append(items, t.translateFeedItem(entry))
	}
	return items
}

func (t *DefaultAtomTranslator) translateItemTitle(entry *atom.Entry) (title string) {
	return entry.Title
}

func (t *DefaultAtomTranslator) translateItemDescription(entry *atom.Entry) (desc string) {
	return entry.Summary
}

func (t *DefaultAtomTranslator) translateItemContent(entry *atom.Entry) (content string) {
	if entry.Content != nil {
		content = entry.Content.Value
	}
	return content
}

func (t *DefaultAtomTranslator) translateItemLink(entry *atom.Entry) (link string) {
	l := t.firstLinkWithType("alternate", entry.Links)
	if l != nil {
		link = l.Href
	}
	return link
}

func (t *DefaultAtomTranslator) translateItemLinks(entry *atom.Entry) (links []string) {
	for _, l := range entry.Links {
		if l.Rel == "" || l.Rel == "alternate" || l.Rel == "self" {
			links = append(links, l.Href)
		}
	}
	return links
}

func (t *DefaultAtomTranslator) translateItemUpdated(entry *atom.Entry) (updated string) {
	return entry.Updated
}

func (t *DefaultAtomTranslator) translateItemUpdatedParsed(entry *atom.Entry) (updated *time.Time) {
	return entry.UpdatedParsed
}

func (t *DefaultAtomTranslator) translateItemPublished(entry *atom.Entry) (published string) {
	published = entry.Published
	if published == "" {
		published = entry.Updated
	}
	return published
}

func (t *DefaultAtomTranslator) translateItemPublishedParsed(entry *atom.Entry) (published *time.Time) {
	published = entry.PublishedParsed
	if published == nil {
		published = entry.UpdatedParsed
	}
	return published
}

func (t *DefaultAtomTranslator) translateItemAuthor(entry *atom.Entry) (author *Person) {
	a := t.firstPerson(entry.Authors)
	if a != nil {
		author = &Person{}
		author.Name = a.Name
		author.Email = a.Email
	}
	return author
}

func (t *DefaultAtomTranslator) translateItemAuthors(entry *atom.Entry) (authors []*Person) {
	if entry.Authors != nil {
		authors = make([]*Person, 0, len(entry.Authors))
		for _, a := range entry.Authors {
			authors = append(authors, &Person{
				Name:  a.Name,
				Email: a.Email,
			})
		}
	}
	return authors
}

func (t *DefaultAtomTranslator) translateItemGUID(entry *atom.Entry) (guid string) {
	return entry.ID
}

func (t *DefaultAtomTranslator) translateItemImage(_ *atom.Entry) (image *Image) {
	return nil
}

func (t *DefaultAtomTranslator) translateItemCategories(entry *atom.Entry) (categories []string) {
	if entry.Categories != nil {
		categories = make([]string, 0, len(entry.Categories))
		for _, c := range entry.Categories {
			if c.Label != "" {
				categories = append(categories, c.Label)
			} else {
				categories = append(categories, c.Term)
			}
		}
	}
	return categories
}

func (t *DefaultAtomTranslator) translateItemEnclosures(entry *atom.Entry) (enclosures []*Enclosure) {
	if entry.Links != nil {
		enclosures = make([]*Enclosure, 0, len(entry.Links))
		for _, e := range entry.Links {
			if e.Rel == "enclosure" {
				enclosure := &Enclosure{}
				enclosure.URL = e.Href
				enclosure.Length = e.Length
				enclosure.Type = e.Type
				enclosures = append(enclosures, enclosure)
			}
		}

		if len(enclosures) == 0 {
			enclosures = nil
		}
	}
	return enclosures
}

func (t *DefaultAtomTranslator) firstLinkWithType(linkType string, links []*atom.Link) *atom.Link {
	if links == nil {
		return nil
	}

	for _, link := range links {
		if link.Rel == linkType {
			return link
		}
	}
	return nil
}

func (t *DefaultAtomTranslator) firstPerson(persons []*atom.Person) (person *atom.Person) {
	if len(persons) == 0 {
		return person
	}
	return persons[0]
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
