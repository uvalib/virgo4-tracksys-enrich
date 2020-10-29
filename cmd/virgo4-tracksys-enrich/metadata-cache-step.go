package main

import (
	"github.com/uvalib/virgo4-sqs-sdk/awssqs"
	"log"
)

// this is our actual implementation
type metadataCacheStepImpl struct {
	s3proxy *S3Proxy // our S3 abstraction
}

// NewMetaDataCacheStep - the factory
func NewMetaDataCacheStep(config *ServiceConfig) PipelineStep {

	// mock implementation here if necessary

	impl := &metadataCacheStepImpl{}
	impl.s3proxy = NewS3Proxy(config)
	return impl
}

func (si *metadataCacheStepImpl) Name() string {
	return "Metadata cache"
}

func (si *metadataCacheStepImpl) Process(message *awssqs.Message, data interface{}) (bool, interface{}, error) {

	tracksysData, ok := data.(TrackSysItemDetails)
	if ok == false {
		log.Printf("ERROR: failed to type assert into known payload")
		return false, data, ErrorTypeAssertion
	}

	err := si.createMetadataCache(tracksysData, message)
	if err != nil {
		return false, data, err
	}

	return true, data, nil
}

func (si *metadataCacheStepImpl) createMetadataCache(tracksysDetails TrackSysItemDetails, message *awssqs.Message) error {

	metadata, err := si.createMetadataContent(tracksysDetails, message)
	if err != nil {
		return err
	}

	key := tracksysDetails.SirsiId
	err = si.s3proxy.WriteToCache(key, metadata)
	if err != nil {
		return err
	}
	return nil
}

func (si *metadataCacheStepImpl) createMetadataContent(tracksysDetails TrackSysItemDetails, message *awssqs.Message) (string, error) {

	metadata := "{\"message\":\"some interesting metadata\"}"
	return metadata, nil
}

//
// end of file
//
