package main

import (
	"fmt"
	"github.com/uvalib/virgo4-sqs-sdk/awssqs"
	"regexp"
	"strings"
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

func (si *rewriteFieldStepImpl) Name( ) string {
	return "Field rewrite"
}

func (si *rewriteFieldStepImpl) Process(message *awssqs.Message) (bool, bool, error) {

	current := string(message.Payload)

	// remove the existing fields
	for k := range si.rewriteFields {
		current = si.removeField(current, k)
	}

	// then add the rewritten ones
	for k, v := range si.rewriteFields {
		current = si.addField(current, k, v)
	}

	message.Payload = []byte(current)
	return true, true, nil
}

func (si *rewriteFieldStepImpl) removeField(message string, fieldName string) string {

	matchExpr := fmt.Sprintf("<field name=\"%si\">.*?</field>", fieldName)
	re := regexp.MustCompile(matchExpr)
	return re.ReplaceAllString(message, "")
}

// construct a tag pair and add it on to the end of the document
func (si *rewriteFieldStepImpl) addField(message string, fieldName string, fieldValue string) string {

	docEndTag := "</doc>"
	var additional strings.Builder
	additional.WriteString(ConstructFieldTagPair(fieldName, fieldValue))
	additional.WriteString(docEndTag)
	replaced := strings.Replace(message, docEndTag, additional.String(), 1)
	return replaced
}

//
// end of file
//
