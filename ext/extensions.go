package ext

import (
	"iter"
)

// Extensions is the generic extension map for Feeds and Items.
// The first map is for the element namespace prefix (e.g., itunes).
// The second map is for the element name (e.g., author).
type Extensions map[string]map[string][]Extension

// Extension represents a single XML element that was in a non
// default namespace in a Feed or Item/Entry.
type Extension struct {
	Name     string                 `json:"name"`
	Value    string                 `json:"value"`
	Attrs    map[string]string      `json:"attrs"`
	Children map[string][]Extension `json:"children"`
}

func ElementsSeq(extensions Extensions, keys ...string,
) iter.Seq[map[string][]Extension] {
	return func(yield func(map[string][]Extension) bool) {
		if extensions == nil {
			return
		}
		for _, key := range keys {
			if match, ok := extensions[key]; ok {
				if !yield(match) {
					return
				}
			}
		}
	}
}
