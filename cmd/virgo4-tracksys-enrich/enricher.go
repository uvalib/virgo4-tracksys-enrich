package main

import (
	"bytes"
	"encoding/xml"
	"net/http"
	//"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/uvalib/virgo4-sqs-sdk/awssqs"
)

// a SOLR limitation
//var maxSolrFieldSize = 32765

//
// Much of this code is based on the existing SOLRMark plugin to handle tracksys decoration. See more details:
// https://github.com/uvalib/utilities/blob/master/bib/bin/solrmarc3/index_java/src/DlMixin.java
// https://github.com/uvalib/utilities/blob/master/bib/bin/solrmarc3/dl_augment.properties
//

var errorNoIdentifier = fmt.Errorf("no identifier attribute located for document")

// our interface
type Enricher interface {
	Enrich(CacheLoader, *awssqs.Message) error
}

// this is our actual implementation
type enrichImpl struct {
	RightsEndpoint string       // the Rights URL
	httpClient     *http.Client // our http client connection
}

// Initialize our SOLR connection
func NewEnricher(config *ServiceConfig) Enricher {

	// mock implementation here if necessary

	impl := &enrichImpl{}

	impl.httpClient = newHttpClient(2, config.ServiceTimeout)
	impl.RightsEndpoint = config.RightsEndpoint

	return impl
}

func (e *enrichImpl) Enrich(cache CacheLoader, message *awssqs.Message) error {

	// extract the ID else we cannot do anything
	id, found := message.GetAttribute(awssqs.AttributeKeyRecordId)
	if found == true {
		found, err := cache.Contains(id)
		if err != nil {
			return err
		}

		// tracksys contains information about this item
		if found == true {

			// we have determined that we do not want to enrich certain classs of item
			toEnrich := e.enrichableItem( message )
			if toEnrich == true {
				log.Printf("INFO: located id %s in tracksys cache, getting details", id)
				trackSysDetails, err := cache.Lookup(id)
				if err != nil {
					return err
				}
				err = e.applyEnrichment(trackSysDetails, message)
				if err != nil {
					return err
				}
			} else {
				log.Printf("INFO: id %s is a special item, ignoring it", id)
			}
		}
	} else {
		log.Printf("ERROR: no identifier attribute located for document, no enrichment possible")
		return errorNoIdentifier
	}

	return nil
}

// there are certain classes of item that should not be enriched, not sure why but at the moment tracksys times
// out when we request them.
func (e *enrichImpl) enrichableItem( message *awssqs.Message ) bool {

	// serch for the "serials" facade field
	facetTag := e.makeFieldTagPair( "pool_f_stored", "serials" )
	if strings.Contains( string( message.Payload) , facetTag ) {
		log.Printf("INFO: found %s in payload", facetTag )
		return false
	}

	return true
}

func (e *enrichImpl) applyEnrichment(tracksysDetails *TrackSysItemDetails, message *awssqs.Message) error {

	// extract the information from the tracksys structure
	format_facets, _ := e.extractFormatFacets(tracksysDetails)
	feature_facets, _ := e.extractFeatureFacets(tracksysDetails)
	source_facets, _ := e.extractSourceFacets(tracksysDetails)
	marc_display_facets := []string{"true"}

	additional_collection_facets, _ := e.extractAdditionalCollectionFacets(tracksysDetails)
	alternate_id_facets, _ := e.extractAlternateIdFacets(tracksysDetails)
	individual_call_number_display, _ := e.extractCallNumbers(tracksysDetails)
	//iiif_presentation_metadata_display, err := e.extractIIIFManifest( tracksysDetails )
	//if err != nil {
	//   return err
	//}
	thumbnail_url_display, _ := e.extractThumbnailUrlDisplay(tracksysDetails)
	rights_wrapper_url_display, _ := e.extractRightsWrapperUrlDisplay(tracksysDetails)
	rights_wrapper_display, _ := e.extractRightsWrapperDisplay(tracksysDetails)
	pdf_url_display, _ := e.extractPdfUrlDisplay(tracksysDetails)
	policy_facets, err := e.extractPolicyFacets(tracksysDetails)
	if err != nil {
		return err
	}
	despined_barcodes_display, _ := e.extractDespinedBarcodesDisplay(tracksysDetails)

	// build our additional tag data
	var additionalTags strings.Builder

	additionalTags.WriteString(e.makeFieldTagSet("format_f_stored", e.xmlEncodeValues(format_facets)))
	additionalTags.WriteString(e.makeFieldTagSet("feature_f_stored", e.xmlEncodeValues(feature_facets)))
	additionalTags.WriteString(e.makeFieldTagSet("source_f_stored", e.xmlEncodeValues(source_facets)))

	additionalTags.WriteString(e.makeFieldTagSet("marc_display_f_stored", e.xmlEncodeValues(marc_display_facets)))
	additionalTags.WriteString(e.makeFieldTagSet("additional_collection_f_stored", e.xmlEncodeValues(additional_collection_facets)))
	additionalTags.WriteString(e.makeFieldTagSet("alternate_id_f_stored", e.xmlEncodeValues(alternate_id_facets)))
	additionalTags.WriteString(e.makeFieldTagSet("individual_call_number_a", e.xmlEncodeValues(individual_call_number_display)))
	additionalTags.WriteString(e.makeFieldTagSet("thumbnail_url_a", e.xmlEncodeValues(thumbnail_url_display)))
	additionalTags.WriteString(e.makeFieldTagSet("rights_wrapper_url_a", e.xmlEncodeValues(rights_wrapper_url_display)))
	additionalTags.WriteString(e.makeFieldTagSet("rights_wrapper_a", e.xmlEncodeValues(rights_wrapper_display)))
	additionalTags.WriteString(e.makeFieldTagSet("pdf_url_a", e.xmlEncodeValues(pdf_url_display)))
	additionalTags.WriteString(e.makeFieldTagSet("policy_f_stored", e.xmlEncodeValues(policy_facets)))
	additionalTags.WriteString(e.makeFieldTagSet("despined_barcodes_a", e.xmlEncodeValues(despined_barcodes_display)))

	// a special case
	//buf := e.makeFieldTagSet( "iiif_presentation_metadata_a", e.xmlEncodeValues( iiif_presentation_metadata_display ) )
	//sz := len( buf )
	//if sz > maxSolrFieldSize {
	//   log.Printf("WARNING: iiif_presentation_metadata_a field exceeds maximum size, ignoring. size %d", sz )
	//} else {
	//   additionalTags.WriteString(buf)
	//}

	// tack it on the end of the document
	docEndTag := "</doc>"
	//log.Printf( "Enrich with [%s]", additionalTags.String() )
	additionalTags.WriteString(docEndTag)
	current := string(message.Payload)
	current = strings.Replace(current, docEndTag, additionalTags.String(), 1)
	message.Payload = []byte(current)
	return nil
}

func (e *enrichImpl) extractFormatFacets(tracksysDetails *TrackSysItemDetails) ([]string, error) {
	res := []string{"Online"}
	return res, nil
}

func (e *enrichImpl) extractFeatureFacets(tracksysDetails *TrackSysItemDetails) ([]string, error) {
	res := make([]string, 0, 5)
	res = append(res, "availability")
	res = append(res, "iiif")
	res = append(res, "dl_metadata")
	res = append(res, "rights_wrapper")
	if len(tracksysDetails.PdfServiceRoot) != 0 {
		res = append(res, "pdf_service")
	}
	return res, nil
}

func (e *enrichImpl) extractSourceFacets(tracksysDetails *TrackSysItemDetails) ([]string, error) {
	res := []string{"UVA Library Digital Repository"}
	return res, nil
}

func (e *enrichImpl) extractAdditionalCollectionFacets(tracksysDetails *TrackSysItemDetails) ([]string, error) {
	res := make([]string, 0, 1)
	if len(tracksysDetails.Collection) != 0 {
		res = append(res, tracksysDetails.Collection)
	}
	return res, nil
}

func (e *enrichImpl) extractAlternateIdFacets(tracksysDetails *TrackSysItemDetails) ([]string, error) {
	res := make([]string, 0, 10)
	for _, i := range tracksysDetails.Items {
		if len(i.Pid) != 0 {
			res = append(res, i.Pid)
		}
	}
	return res, nil
}

func (e *enrichImpl) extractCallNumbers(tracksysDetails *TrackSysItemDetails) ([]string, error) {
	res := make([]string, 0, 10)
	for _, i := range tracksysDetails.Items {
		if len(i.CallNumber) != 0 {
			res = append(res, i.CallNumber)
		}
	}
	return res, nil
}

func (e *enrichImpl) extractIIIFManifest(tracksysDetails *TrackSysItemDetails) ([]string, error) {

	urls := make([]string, 0, 10)
	for _, i := range tracksysDetails.Items {
		if len(i.BackendIIIFManifestUrl) != 0 {
			urls = append(urls, i.BackendIIIFManifestUrl)
		}
	}

	res := make([]string, 0, 10)
	for _, i := range urls {
		body, err := httpGet(i, e.httpClient)
		if err == nil {
			res = append(res, string(body))
		} else {
			log.Printf("ERROR: endpoint %s returns %s", i, err)
			return nil, err
		}
	}

	return res, nil
}

func (e *enrichImpl) extractThumbnailUrlDisplay(tracksysDetails *TrackSysItemDetails) ([]string, error) {
	res := make([]string, 0, 10)
	for _, i := range tracksysDetails.Items {
		if len(i.ThumbnailUrl) != 0 {
			res = append(res, i.ThumbnailUrl)
		}
	}
	return res, nil
}

func (e *enrichImpl) extractRightsWrapperUrlDisplay(tracksysDetails *TrackSysItemDetails) ([]string, error) {
	res := make([]string, 0, 10)
	for _, i := range tracksysDetails.Items {
		if len(i.RightsWrapperUrl) != 0 {
			res = append(res, i.RightsWrapperUrl)
		}
	}
	return res, nil
}

func (e *enrichImpl) extractRightsWrapperDisplay(tracksysDetails *TrackSysItemDetails) ([]string, error) {
	res := make([]string, 0, 10)
	for _, i := range tracksysDetails.Items {
		if len(i.RightsWrapperText) != 0 {
			res = append(res, i.RightsWrapperText)
		}
	}
	return res, nil
}

func (e *enrichImpl) extractPdfUrlDisplay(tracksysDetails *TrackSysItemDetails) ([]string, error) {
	res := make([]string, 0, 1)
	if len(tracksysDetails.PdfServiceRoot) != 0 {
		res = append(res, tracksysDetails.PdfServiceRoot)
	}
	return res, nil
}

func (e *enrichImpl) extractPolicyFacets(tracksysDetails *TrackSysItemDetails) ([]string, error) {

	res := make([]string, 0, 1)
	for _, i := range tracksysDetails.Items {
		if len(i.Pid) != 0 {
			url := fmt.Sprintf("%s/%s", e.RightsEndpoint, i.Pid)
			body, err := httpGet(url, e.httpClient)
			if err == nil {
				if string(body) != "public" {
					res = append(res, string(body))
				}
				break
			} else {
				log.Printf("ERROR: endpoint %s returns %s", url, err)
				return nil, err
			}
		}
	}

	return res, nil
}

func (e *enrichImpl) extractDespinedBarcodesDisplay(tracksysDetails *TrackSysItemDetails) ([]string, error) {

	res := make([]string, 0, 10)
	if tracksysDetails.Collection == "Gannon Collection" {
		for _, i := range tracksysDetails.Items {
			if len(i.Barcode) != 0 {
				res = append(res, i.Barcode)
			}
		}
	}
	return res, nil
}

func (e *enrichImpl) xmlEncodeValues(values []string) []string {
	for ix, _ := range values {
		values[ix] = e.xmlEscape(values[ix])
	}

	return values
}

func (e *enrichImpl) makeFieldTagSet(name string, values []string) string {
	var res strings.Builder
	for _, v := range values {
		res.WriteString(e.makeFieldTagPair(name, v))
	}
	return res.String()
}

func (e *enrichImpl) makeFieldTagPair(name string, value string) string {
	return fmt.Sprintf("<field name=\"%s\">%s</field>", name, value)
}

func (e *enrichImpl) xmlEscape(value string) string {
	var escaped bytes.Buffer
	_ = xml.EscapeText(&escaped, []byte(value))
	return escaped.String()
}

//
// end of file
//
