package main

import (
	"fmt"
	"github.com/uvalib/virgo4-sqs-sdk/awssqs"
	"log"
)

var errorNoIdentifier = fmt.Errorf("no identifier attribute located for document")

// newly published items may appear here and in this case, the item identifier might not exist
// in a service cache so this attribute is used to ignore the cache and lookup the item in an external
// service anyway.
var ignoreCacheAttributeName = "ignore-cache"

// this is our actual implementation
type tracksysExtractStepImpl struct {
}

// NewTracksysExtractStep - the factory
func NewTracksysExtractStep(config *ServiceConfig) PipelineStep {

	// mock implementation here if necessary

	impl := &tracksysExtractStepImpl{}
	return impl
}

func (si *tracksysExtractStepImpl) Name() string {
	return "Tracksys extract"
}

func (si *tracksysExtractStepImpl) Process(message *awssqs.Message, _ interface{}) (bool, interface{}, error) {

	// extract the ID else we cannot do anything
	id, foundId := message.GetAttribute(awssqs.AttributeKeyRecordId)
	if foundId == true {

		var err error
		var lookupTrackSys = false

		// see if we have the attribute telling us to ignore the cache
		_, foundId = message.GetAttribute(ignoreCacheAttributeName)
		if foundId == true {
			log.Printf("INFO: id %s marked to IGNORE tracksys cache, getting details", id)
			lookupTrackSys = true
		} else {
			// look the item up in the cache to see if tracksys knows about it
			lookupTrackSys, err = TracksysIdCache.Contains(id)
			if err != nil {
				return false, nil, err
			}
			if lookupTrackSys == true {
				log.Printf("INFO: located id %s in tracksys cache, getting details", id)
			}
		}

		// tracksys (probably) contains information about this item
		if lookupTrackSys == true {
			// actually do the lookup work
			trackSysDetails, err := TracksysIdCache.Lookup(id)
			if err != nil {
				return false, nil, err
			}
			// we found the item in tracksys
			return true, *trackSysDetails, nil
		} else {
			//log.Printf("DEBUG: %s is not in the cache, no further processing", id )
			// item is not in tracksys, no further processing required
			return false, nil, nil
		}
	}

	log.Printf("ERROR: no identifier attribute located for document, no tracksys lookup possible")
	return false, nil, errorNoIdentifier
}

//
// end of file
//
