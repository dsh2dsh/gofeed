package date

import (
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// rfc822Zones maps RFC 822 §5 timezone abbreviations to their corresponding
// time.Location. Go's time.Parse recognises the abbreviation but assigns offset
// 0, and time.LoadLocation rejects short names such as "EDT", so we need an
// explicit table.
var rfc822Zones = map[string]*time.Location{
	// included for completeness; offset-0 entries are no-ops in fixRFC822Zone
	"UT":  time.FixedZone("UT", 0),
	"GMT": time.FixedZone("GMT", 0),
	"EST": time.FixedZone("EST", -5*3600),
	"EDT": time.FixedZone("EDT", -4*3600),
	"CST": time.FixedZone("CST", -6*3600),
	"CDT": time.FixedZone("CDT", -5*3600),
	"MST": time.FixedZone("MST", -7*3600),
	"MDT": time.FixedZone("MDT", -6*3600),
	"PST": time.FixedZone("PST", -8*3600),
	"PDT": time.FixedZone("PDT", -7*3600),
}

func locationFromCache(tz string) *time.Location {
	return cachedLocations.Location(tz)
}

type locationCache struct {
	mu        sync.RWMutex
	locations map[string]*time.Location
	sg        singleflight.Group
}

var cachedLocations = &locationCache{locations: make(map[string]*time.Location)}

func (self *locationCache) Location(tz string) *time.Location {
	if rfc822Loc, ok := rfc822Zones[tz]; ok {
		return rfc822Loc
	}

	self.mu.RLock()
	loc, ok := self.locations[tz]
	self.mu.RUnlock()
	if ok {
		return loc
	}

	v, _, _ := self.sg.Do(tz, func() (any, error) {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			loc = time.UTC
		}

		self.mu.Lock()
		self.locations[tz] = loc
		self.mu.Unlock()
		return loc, nil
	})
	return v.(*time.Location)
}
