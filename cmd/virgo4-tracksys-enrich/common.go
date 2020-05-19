package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
)

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
