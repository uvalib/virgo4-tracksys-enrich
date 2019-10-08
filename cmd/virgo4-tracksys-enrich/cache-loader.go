package main

import (
   "fmt"
   "net/http"
   "time"
)

// our interface
type CacheLoader interface {
   Contains( string ) ( bool, error )
}

// this is our actual implementation
type cacheLoaderImpl struct {

   directoryUrl    string           // the URL for requesting a directory of contents
   detailsUrl      string           // the URL for requesting details
   httpClient    * http.Client      // our http client connection

   cacheImpl       Cache            // the actual cache
   cacheLoaded     time.Time        // when we last repopulated the cache
   cacheMaxAge     time.Duration    // the maximum age of the cache
}

// Initialize our cache loader
func NewCacheLoader( config * ServiceConfig ) ( CacheLoader, error ) {

   // mock implementation here if necessary

   cache := NewCache()
   loader := &cacheLoaderImpl{ cacheImpl: cache }
   loader.directoryUrl = fmt.Sprintf( "%s/%s", config.ServiceEndpoint, config.ApiDirectoryPath )
   loader.detailsUrl = fmt.Sprintf( "%s/%s", config.ServiceEndpoint, config.ApiDetailsPath )
   loader.cacheMaxAge = time.Duration ( config.CacheAge ) * time.Second

   // configure the client
   loader.httpClient = &http.Client {
      Transport: &http.Transport{
         MaxIdleConnsPerHost: 5,
      },
      Timeout: time.Duration( config.ServiceTimeout ) * time.Second,
   }

   // reload the cache
   err := loader.reload( )
   return loader, err
}

// lookup in the cache, refresh as necessary
func ( cl * cacheLoaderImpl ) Contains( id string ) ( bool, error ) {

   if cl.cacheStale( ) == true {
      err := cl.reload( )
      if err != nil {
         return false, err
      }
   }

   return cl.cacheImpl.Contains( id ), nil
}

// reload the cache
func ( cl * cacheLoaderImpl ) reload( ) error {

   contents, err := cl.protocolDirectory( cl.directoryUrl )
   if err != nil {
      return err
   }

   // reload the cache
   cl.cacheImpl.Reload( contents )
   cl.cacheLoaded = time.Now( )

   return nil
}

func ( cl * cacheLoaderImpl ) cacheStale( ) bool {

   duration := time.Since( cl.cacheLoaded )
   return duration.Seconds( ) > cl.cacheMaxAge.Seconds( )
}

//
// end of file
//
