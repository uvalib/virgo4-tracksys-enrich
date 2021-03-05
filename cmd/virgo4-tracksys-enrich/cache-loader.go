package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// CacheLoader - our interface
type CacheLoader interface {
	Contains(string) (bool, error)
	Lookup(string) (*TracksysSirsiItem, error)
}

// TracksysIdCache our singleton store
var TracksysIdCache CacheLoader

// this is our actual implementation
type cacheLoaderImpl struct {
	loadApi    string // the API path for loading the cache list
	detailsApi string // the API path for requesting details for items in the cache
	pidApi     string // the API path for requesting PID details
	multiMode  bool   // do we expect single or m ultiple items from the endpoint

	httpClient *http.Client // our http client connection

	cacheImpl   Cache         // the actual cache
	cacheLoaded time.Time     // when we last repopulated the cache
	cacheMaxAge time.Duration // the maximum age of the cache

	mu sync.RWMutex // coordinate cache reloads
}

// NewCacheLoader - the factory
func NewCacheLoader(config *ServiceConfig) error {

	// mock implementation here if necessary

	cache := NewCache()
	impl := &cacheLoaderImpl{cacheImpl: cache}
	impl.multiMode = config.Mode == "sirsi"
	impl.loadApi = fmt.Sprintf("%s/%s", config.ServiceEndpoint, config.CacheLoadApi)
	impl.detailsApi = fmt.Sprintf("%s/%s", config.ServiceEndpoint, config.CacheDetailsApi)
	impl.pidApi = fmt.Sprintf("%s/%s", config.ServiceEndpoint, config.PidDetailsApi)
	impl.cacheMaxAge = time.Duration(config.CacheAge) * time.Second

	// configure the http client
	impl.httpClient = newHttpClient(config.Workers, config.ServiceTimeout)

	// assign to our global singleton
	TracksysIdCache = impl

	// reload the cache
	return impl.reload()
}

// Contains - lookup in the cache, refresh as necessary
func (cl *cacheLoaderImpl) Contains(id string) (bool, error) {

	if cl.cacheStale() == true {

		// lock while we refresh the cache
		cl.mu.Lock()
		defer cl.mu.Unlock()

		// double check pattern
		if cl.cacheStale() == true {
			log.Printf("INFO: cache is stale, time to reload")
			err := cl.reload()
			if err != nil {
				return false, err
			}
		}
	}

	return cl.cacheImpl.Contains(id), nil
}

// Lookup - lookup an item... we know (or think we know) it exists so we
// get the details from Tracksys
func (cl *cacheLoaderImpl) Lookup(id string) (*TracksysSirsiItem, error) {

	var tsItem *TracksysSirsiItem
	var err error

	if cl.multiMode == true {
		// we expect sirsi items from the details API
		tsItem, err = cl.protocolGetSirsiDetails(fmt.Sprintf("%s/%s", cl.detailsApi, id))
		if err != nil {
			return nil, err
		}

		var pidItem *TracksysPidItem
		// for each of the parts in the item
		for ix, part := range tsItem.Items {
			// get some PID details so we can determine if this is an OCR candidate
			pidItem, err = cl.protocolGetPidDetails(fmt.Sprintf("%s/%s", cl.pidApi, part.Pid))
			if err != nil {
				return nil, err
			}

			tsItem.Items[ix].OcrCandidate = pidItem.OcrCandidate
		}

	} else {
		// we expect image (partial) items from the details API
		var tsPart *TracksysPart
		tsPart, err = cl.protocolGetImageDetails(fmt.Sprintf("%s/%s", cl.detailsApi, id))
		if err != nil {
			return nil, err
		}
		item := []TracksysPart{*tsPart}
		tsItem = &TracksysSirsiItem{Items: item}
	}

	return tsItem, nil
}

// reload the cache
func (cl *cacheLoaderImpl) reload() error {

	contents, err := cl.protocolGetKnownIds(cl.loadApi)

	// after discussions with Mike, we determined that failing when attempting to reload the cache is a fatal set of
	// circumstances and we should not continue to process items
	fatalIfError(err)

	// reload the cache
	cl.cacheImpl.Reload(contents.Items)
	cl.cacheLoaded = time.Now()

	return nil
}

func (cl *cacheLoaderImpl) cacheStale() bool {

	duration := time.Since(cl.cacheLoaded)
	return duration.Seconds() > cl.cacheMaxAge.Seconds()
}

//
// end of file
//
