package main

import (
	"github.com/patrickmn/go-cache"
	"log"
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
	for ix, _ := range ids {

		// Tracksys returns empty ID's and duplicate ID's so fix this here...
		if len( ids[ix] ) != 0 {
			_, found := ci.c.Get(ids[ix])
			if found == false {
				//fmt.Printf("[%s]\n", ids[ix] )
				ci.c.Set(ids[ix], 0, cache.NoExpiration)
			}
		}
	}

	log.Printf("Loaded cache with %d items", ci.c.ItemCount())
}

//
// does the supplied id exist in the cache
//
func (ci *cacheImpl) Contains(id string) bool {

	// lookup the id in the cache
	_, found := ci.c.Get(id)
	if found {
		//log.Printf( "ID [%s] found", id )
		return true
	}

	return false
}

//
// end of file
//
