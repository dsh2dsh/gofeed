package atom

import "strings"

type textAttributes struct {
	Type string
	Mode string
}

func (self *textAttributes) XHTML() bool {
	return self.Type == "xhtml" || strings.Contains(self.Type, "/xhtml")
}

func (self *textAttributes) XML() bool {
	return self.Mode == "xml" || strings.HasSuffix(self.Type, "+xml") ||
		strings.HasSuffix(self.Type, "/xml")
}

func (self *textAttributes) Encoded() bool {
	if self.Mode == "base64" {
		return true
	}

	textEncoding := self.Type == "text" || self.Type == "html" ||
		strings.HasPrefix(self.Type, "text/") ||
		(self.Type == "" && self.Mode == "")
	return !textEncoding
}
