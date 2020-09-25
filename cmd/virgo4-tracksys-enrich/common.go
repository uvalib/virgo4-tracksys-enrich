package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
)

var ErrorTypeAssertion = fmt.Errorf("failure to type assert interface")

// construct a tag pair and append it on to the end of the document
func AppendXmlField(message string, fieldName string, fieldValue string) string {

	docEndTag := "</doc>"
	var additional strings.Builder
	additional.WriteString(ConstructFieldTagPair(fieldName, fieldValue))
	additional.WriteString(docEndTag)
	replaced := strings.Replace(message, docEndTag, additional.String(), 1)
	return replaced
}

// remove the specified field from the document
func RemoveXmlField(message string, fieldName string) string {

	matchExpr := fmt.Sprintf("<field name=\"%s\">.*?</field>", fieldName)
	re := regexp.MustCompile(matchExpr)
	return re.ReplaceAllString(message, "")
}

// extract multiple field values from the document
func ExtractXmlFields(message string, fieldName string) []string {

	matchExpr := fmt.Sprintf("<field name=\"%s\">(.*?)</field>", fieldName)
	re := regexp.MustCompile(matchExpr)
	rs := re.FindAllStringSubmatch(message, -1)
	// no match
	if rs == nil {
		return []string{}
	}
	// return the matches
	var res []string
	for _, m := range rs {
		if len(m[1]) != 0 {
			res = append(res, m[1])
		}
	}
	return res
}

func XmlEncodeValues(values []string) []string {
	for ix := range values {
		values[ix] = xmlEscape(values[ix])
	}

	return values
}

func ConstructFieldTagSet(name string, values []string) string {
	var res strings.Builder
	for _, v := range values {
		res.WriteString(ConstructFieldTagPair(name, v))
	}
	return res.String()
}

func ConstructFieldTagPair(name string, value string) string {
	return fmt.Sprintf("<field name=\"%s\">%s</field>", name, value)
}

func xmlEscape(value string) string {
	var escaped bytes.Buffer
	_ = xml.EscapeText(&escaped, []byte(value))
	return escaped.String()
}

//
// end of file
//
