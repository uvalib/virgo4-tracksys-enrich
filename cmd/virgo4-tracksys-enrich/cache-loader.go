package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// CacheLoader - our interface
type CacheLoader interface {
	Contains(string) (bool, error)
	Lookup(string) (*TrackSysItemDetails, error)
}

// TracksysIdCache our singleton store
var TracksysIdCache CacheLoader

// this is our actual implementation
type cacheLoaderImpl struct {
	directoryUrl string       // the URL for requesting a directory of contents
	detailsUrl   string       // the URL for requesting details
	httpClient   *http.Client // our http client connection

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
	impl.directoryUrl = fmt.Sprintf("%s/%s", config.ServiceEndpoint, config.ApiDirectoryPath)
	impl.detailsUrl = fmt.Sprintf("%s/%s", config.ServiceEndpoint, config.ApiDetailsPath)
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

// Lookup - lookup an item in the cache
func (cl *cacheLoaderImpl) Lookup(id string) (*TrackSysItemDetails, error) {

	details, err := cl.protocolDetails(fmt.Sprintf("%s/%s", cl.detailsUrl, id))
	if err != nil {
		return nil, err
	}

	ts, err := cl.decodeTracksysPayload(details)
	if err != nil {
		return nil, err
	}

	return ts, nil
}

// reload the cache
func (cl *cacheLoaderImpl) reload() error {

	contents, err := cl.protocolDirectory(cl.directoryUrl)

	// after discussions with Mike, we determined that failing when attempting to reload the cache is a fatal set of
	// circumstances and we should not continue to process items
	fatalIfError(err)

	// reload the cache
	cl.cacheImpl.Reload(contents)
	cl.cacheLoaded = time.Now()

	return nil
}

func (cl *cacheLoaderImpl) cacheStale() bool {

	duration := time.Since(cl.cacheLoaded)
	return duration.Seconds() > cl.cacheMaxAge.Seconds()
}

// decode the Tracksys payload from the supplied payload
func (cl *cacheLoaderImpl) decodeTracksysPayload(payload []byte) (*TrackSysItemDetails, error) {

	td := TrackSysItemDetails{}
	err := json.Unmarshal(payload, &td)
	if err != nil {
		log.Printf("ERROR: json unmarshal: %s", err)
		return nil, err
	}
	return &td, nil
}

//
// end of file
//
