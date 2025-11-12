package ext

import "iter"

// https://www.rssboard.org/media-rss
type Media struct {
	Groups   []MediaGroup   `json:"group,omitempty"`
	Contents []MediaContent `json:"content,omitempty"`
}

type MediaGroup struct {
	Categories   []string           `json:"category,omitempty"`
	Contents     []MediaContent     `json:"content,omitempty"`
	Thumbnails   []string           `json:"thumbnail,omitempty"`
	Titles       []MediaDescription `json:"title,omitempty"`
	Descriptions []MediaDescription `json:"description,omitempty"`
	PeerLinks    []MediaPeerLink    `json:"peerLink,omitempty"`
}

type MediaContent struct {
	URL      string `json:"url,omitempty"`
	Type     string `json:"type,omitempty"`
	FileSize string `json:"fileSize,omitempty"`
	Medium   string `json:"medium,omitempty"`

	Categories   []string           `json:"category,omitempty"`
	Thumbnails   []string           `json:"thumbnail,omitempty"`
	Titles       []MediaDescription `json:"title,omitempty"`
	Descriptions []MediaDescription `json:"description,omitempty"`
	PeerLinks    []MediaPeerLink    `json:"peerLink,omitempty"`
}

type MediaDescription struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
}

type MediaPeerLink struct {
	URL  string `json:"href,omitempty"`
	Type string `json:"type,omitempty"`
}

func (self *Media) AllCategories() iter.Seq[string] {
	return self.categoriesIter
}

func (self *Media) categoriesIter(yield func(string) bool) {
	for _, c := range self.Contents {
		for _, s := range c.Categories {
			if !yield(s) {
				return
			}
		}
	}

	for _, g := range self.Groups {
		for s := range g.AllCategories() {
			if !yield(s) {
				return
			}
		}
	}
}

func (self *Media) AllContents() iter.Seq[MediaContent] {
	return self.contentsIter
}

func (self *Media) contentsIter(yield func(MediaContent) bool) {
	for _, c := range self.Contents {
		if !yield(c) {
			return
		}
	}

	for _, g := range self.Groups {
		for _, c := range g.Contents {
			if !yield(c) {
				return
			}
		}
	}
}

func (self *Media) AllPeerLinks() iter.Seq[MediaPeerLink] {
	return self.peerLinksIter
}

func (self *Media) peerLinksIter(yield func(MediaPeerLink) bool) {
	for _, c := range self.Contents {
		for _, pl := range c.PeerLinks {
			if pl.URL != "" && !yield(pl) {
				return
			}
		}
	}

	for _, g := range self.Groups {
		for pl := range g.AllPeerLinks() {
			if !yield(pl) {
				return
			}
		}
	}
}

func (self *Media) AllThumbnails() iter.Seq[string] {
	return self.thumbnailsIter
}

func (self *Media) thumbnailsIter(yield func(string) bool) {
	for _, c := range self.Contents {
		for _, s := range c.Thumbnails {
			if s != "" && !yield(s) {
				return
			}
		}
	}

	for _, g := range self.Groups {
		for s := range g.AllThumbnails() {
			if !yield(s) {
				return
			}
		}
	}
}

func (self *Media) Description() string {
	for _, c := range self.Contents {
		for _, d := range c.Descriptions {
			if d.Type == "html" {
				return d.Text
			}
		}
	}

	for _, g := range self.Groups {
		for _, d := range g.Descriptions {
			if d.Type == "html" {
				return d.Text
			}
		}
		for _, c := range g.Contents {
			for _, d := range c.Descriptions {
				if d.Type == "html" {
					return d.Text
				}
			}
		}
	}
	return ""
}

func (self *MediaGroup) AllCategories() iter.Seq[string] {
	return self.categoriesIter
}

func (self *MediaGroup) categoriesIter(yield func(string) bool) {
	for _, s := range self.Categories {
		if !yield(s) {
			return
		}
	}

	for _, c := range self.Contents {
		for _, s := range c.Categories {
			if !yield(s) {
				return
			}
		}
	}
}

func (self *MediaGroup) AllPeerLinks() iter.Seq[MediaPeerLink] {
	return self.peerLinksIter
}

func (self *MediaGroup) peerLinksIter(yield func(MediaPeerLink) bool) {
	for _, pl := range self.PeerLinks {
		if pl.URL != "" && !yield(pl) {
			return
		}
	}

	for _, c := range self.Contents {
		for _, pl := range c.PeerLinks {
			if pl.URL != "" && !yield(pl) {
				return
			}
		}
	}
}

func (self *MediaGroup) AllThumbnails() iter.Seq[string] {
	return self.thumbnailsIter
}

func (self *MediaGroup) thumbnailsIter(yield func(string) bool) {
	for _, s := range self.Thumbnails {
		if s != "" && !yield(s) {
			return
		}
	}

	for _, c := range self.Contents {
		for _, s := range c.Thumbnails {
			if s != "" && !yield(s) {
				return
			}
		}
	}
}
