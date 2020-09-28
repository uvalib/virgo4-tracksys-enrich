package main

import (
	"fmt"
	"github.com/uvalib/virgo4-sqs-sdk/awssqs"
	"log"
)

var errorNoIdentifier = fmt.Errorf("no identifier attribute located for document")

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
		inTrackSys, err := TracksysIdCache.Contains(id)
		if err != nil {
			return false, nil, err
		}

		// tracksys contains information about this item
		if inTrackSys == true {

			log.Printf("INFO: located id %s in tracksys cache, getting details", id)
			trackSysDetails, err := TracksysIdCache.Lookup(id)
			if err != nil {
				return false, nil, err
			}
			// we found the item in tracksys
			return true, *trackSysDetails, nil
		} else {
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
