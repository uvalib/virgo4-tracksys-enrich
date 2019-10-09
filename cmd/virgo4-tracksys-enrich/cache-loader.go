package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// our interface
type CacheLoader interface {
	Contains(string) (bool, error)
	Lookup(string) (* TrackSysItemDetails, error)
}

// this is our actual implementation
type cacheLoaderImpl struct {
	directoryUrl string       // the URL for requesting a directory of contents
	detailsUrl   string       // the URL for requesting details
	httpClient   *http.Client // our http client connection

	cacheImpl   Cache         // the actual cache
	cacheLoaded time.Time     // when we last repopulated the cache
	cacheMaxAge time.Duration // the maximum age of the cache

	mu          sync.RWMutex  // coordinate cache reloads
}

// Initialize our cache loader
func NewCacheLoader(config *ServiceConfig) (CacheLoader, error) {

	// mock implementation here if necessary

	cache := NewCache()
	loader := &cacheLoaderImpl{cacheImpl: cache}
	loader.directoryUrl = fmt.Sprintf("%s/%s", config.ServiceEndpoint, config.ApiDirectoryPath)
	loader.detailsUrl = fmt.Sprintf("%s/%s", config.ServiceEndpoint, config.ApiDetailsPath)
	loader.cacheMaxAge = time.Duration(config.CacheAge) * time.Second

	// configure the client
	loader.httpClient = &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 5,
		},
		Timeout: time.Duration(config.ServiceTimeout) * time.Second,
	}

	// reload the cache
	err := loader.reload()
	return loader, err
}

// lookup in the cache, refresh as necessary
func (cl *cacheLoaderImpl) Contains(id string) (bool, error) {

	if cl.cacheStale() == true {

		// lock while we refresh the cache
		cl.mu.Lock()
		defer cl.mu.Unlock()

		// double check pattern
		if cl.cacheStale() == true {
			err := cl.reload()
			if err != nil {
				return false, err
			}
		}
	}

	return cl.cacheImpl.Contains(id), nil
}

func (cl *cacheLoaderImpl) Lookup(id string) (* TrackSysItemDetails, error) {

	details, err := cl.protocolDetails( fmt.Sprintf( "%s/%s", cl.detailsUrl, id ) )
	if err != nil {
		return nil, err
	}

	ts, err := cl.decodeTracksysPayload( details )
	if err != nil {
		return nil, err
	}

	return ts, nil
}

// reload the cache
func (cl *cacheLoaderImpl) reload() error {

	contents, err := cl.protocolDirectory(cl.directoryUrl)
	if err != nil {
		return err
	}

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
func (cl *cacheLoaderImpl) decodeTracksysPayload(payload []byte) (* TrackSysItemDetails, error) {

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
