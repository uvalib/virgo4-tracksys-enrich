package main

import (
	"github.com/patrickmn/go-cache"
)

// the cache contains a series of identifiers taken from an external system, no other content is cached and individual cache
// entries are not expired. The cache is reloaded on a regular basis

type Cache interface {
	Reload([]string)
	Contains(string) bool
}

// our implementation
type cacheImpl struct {
	c *cache.Cache
}

//
// factory
//
func NewCache() Cache {

	impl := &cacheImpl{}
	impl.c = cache.New(cache.NoExpiration, cache.NoExpiration)
	return impl
}

//
// reload the cache from the list of id's provided
//
func (ci *cacheImpl) Reload(ids []string) {

	// clear the cache
	ci.c.Flush()

	// add the ids to the local cache
	for _, id := range ids {
		ci.c.Set(id, "", cache.NoExpiration)
	}
}

//
// does the supplied id exist in the cache
//
func (ci *cacheImpl) Contains(id string) bool {

	// lookup the id in the cache
	_, found := ci.c.Get(id)
	if found {
		//log.Printf( "ID [%s] found in cache", id )
		return true
	}

	return false
}

//
// end of file
//
