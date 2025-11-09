package ext

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
