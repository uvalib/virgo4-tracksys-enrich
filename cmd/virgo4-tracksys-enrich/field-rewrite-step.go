package main

import (
	"github.com/uvalib/virgo4-sqs-sdk/awssqs"
)

// this is our actual implementation
type rewriteFieldStepImpl struct {
	rewriteFields map[string]string // the fields to rewrite and their rewritten values
}

// NewFieldRewriteStep - the factory
func NewFieldRewriteStep(config *ServiceConfig) PipelineStep {

	// mock implementation here if necessary

	impl := &rewriteFieldStepImpl{}
	impl.rewriteFields = config.RewriteFields

	return impl
}

func (si *rewriteFieldStepImpl) Name() string {
	return "Field rewrite"
}

func (si *rewriteFieldStepImpl) Process(message *awssqs.Message, data interface{}) (bool, interface{}, error) {

	current := string(message.Payload)

	// remove the existing fields
	for k := range si.rewriteFields {
		current = RemoveXmlField(current, k)
	}

	// then add the rewritten ones
	for k, v := range si.rewriteFields {
		current = AppendXmlField(current, k, v)
	}

	message.Payload = []byte(current)
	return true, data, nil
}

//
// end of file
//
