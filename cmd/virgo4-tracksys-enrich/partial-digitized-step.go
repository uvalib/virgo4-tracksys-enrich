package main

import (
	"github.com/uvalib/virgo4-sqs-sdk/awssqs"
)

// this is our actual implementation
type partialDigitizedStepImpl struct {
	//rewriteFields map[string]string // the fields to rewrite and their rewritten values
}

// NewPartialDigitizedStep - the factory
func NewPartialDigitizedStep(config *ServiceConfig) PipelineStep {

	// mock implementation here if necessary

	impl := &partialDigitizedStepImpl{}
	//impl.rewriteFields = config.RewriteFields

	return impl
}

func (si *partialDigitizedStepImpl) Name( ) string {
	return "Partial digitized"
}

func (si *partialDigitizedStepImpl) Process(message *awssqs.Message, data interface{}) (bool, interface{}, error) {

	return true, data, nil
}

//
// end of file
//
