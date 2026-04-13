package date

import (
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

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
