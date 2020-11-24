package main

import (
	"github.com/uvalib/virgo4-sqs-sdk/awssqs"
	"log"
)

// BarcodeFieldName the name of the field we are interested in
var BarcodeFieldName = "barcode_e_stored"

// PartiallyDigitizedFieldName the name of the field to add if appropriate
var PartiallyDigitizedFieldName = "digitized_f_stored"

// PartiallyDigitizedFieldValue the value of the field to add if appropriate
var PartiallyDigitizedFieldValue = "partial"

//var errorNoBarcodes = fmt.Errorf("failed to extract barcode fields")

// this is our actual implementation
type partialDigitizedStepImpl struct {
	//rewriteFields map[string]string // the fields to rewrite and their rewritten values
}

// NewPartialDigitizedStep - the factory
func NewPartialDigitizedStep(config *ServiceConfig) PipelineStep {

	// mock implementation here if necessary

	impl := &partialDigitizedStepImpl{}
	return impl
}

func (si *partialDigitizedStepImpl) Name() string {
	return "Partial digitized"
}

func (si *partialDigitizedStepImpl) Process(message *awssqs.Message, data interface{}) (bool, interface{}, error) {

	current := string(message.Payload)

	barcodes := ExtractXmlFields(current, BarcodeFieldName)
	if len(barcodes) != 0 {
		log.Printf("INFO: extracted %d barcode field(s)", len(barcodes))

		tracksysData, ok := data.(TracksysSirsiItem)
		if ok == false {
			log.Printf("ERROR: failed to type assert into known payload")
			return false, data, ErrorTypeAssertion
		}

		digitizedObjectCount := len(tracksysData.Items)
		log.Printf("INFO: tracksys reports %d digitized item(s)", digitizedObjectCount)

		// we have more items than records of digital items so this should be tagged
		if len(barcodes) != digitizedObjectCount {
			log.Printf("INFO: marking as partially digitized")
			current = AppendXmlField(current, PartiallyDigitizedFieldName, PartiallyDigitizedFieldValue)
			message.Payload = []byte(current)
		}

		return true, data, nil
	}

	// we are now processing items without barcodes so this error case is not
	// terminal to the pipeline
	log.Printf("WARNING: failed to extract barcode fields")
	return true, data, nil
}

//
// end of file
//
