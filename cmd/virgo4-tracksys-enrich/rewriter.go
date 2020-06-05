package main

import (
	"fmt"
	"github.com/uvalib/virgo4-sqs-sdk/awssqs"
	"regexp"
	"strings"
)

// Rewriter - the interface
type Rewriter interface {
	Rewrite(*awssqs.Message) error
}

// this is our actual implementation
type rewriteImpl struct {
	rewriteFields map[string]string // the fields to rewrite and their rewritten values
}

// NewRewriter - the factory
func NewRewriter(config *ServiceConfig) Rewriter {

	// mock implementation here if necessary

	impl := &rewriteImpl{}
	impl.rewriteFields = config.RewriteFields

	return impl
}

func (r *rewriteImpl) Rewrite(message *awssqs.Message) error {

	current := string(message.Payload)

	// remove the existing fields
	for k := range r.rewriteFields {
		current = r.removeField(current, k)
	}

	// then add the rewritten ones
	for k, v := range r.rewriteFields {
		current = r.addField(current, k, v)
	}

	message.Payload = []byte(current)
	return nil
}

func (r *rewriteImpl) removeField(message string, fieldName string) string {

	matchExpr := fmt.Sprintf("<field name=\"%s\">.*?</field>", fieldName)
	re := regexp.MustCompile(matchExpr)
	return re.ReplaceAllString(message, "")
}

// construct a tag pair and add it on to the end of the document
func (r *rewriteImpl) addField(message string, fieldName string, fieldValue string) string {

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
