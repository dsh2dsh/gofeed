package atom

import "strings"

type textAttributes struct {
	Type string
	Mode string
}

func (self *textAttributes) XHTML() bool {
	return self.Type == "xhtml" || strings.Contains(self.Type, "/xhtml")
}

func (self *textAttributes) Encoded() bool {
	textEncoding := self.Type == "text" || self.Type == "html" ||
		strings.HasPrefix(self.Type, "text/") ||
		(self.Type == "" && self.Mode == "")
	return !textEncoding
}
