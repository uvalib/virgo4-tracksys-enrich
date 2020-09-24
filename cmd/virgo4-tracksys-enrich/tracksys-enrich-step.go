package main

import (
	"fmt"
	"log"
	"net/http"
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

// this is our actual implementation
type tracksysEnrichStepImpl struct {
	RightsEndpoint string       // the Rights URL
	httpClient     *http.Client // our http client connection
}

// NewTracksysEnrichStep - the factory
func NewTracksysEnrichStep(config *ServiceConfig) PipelineStep {

	// mock implementation here if necessary

	impl := &tracksysEnrichStepImpl{}

	impl.httpClient = newHttpClient(2, config.ServiceTimeout)
	impl.RightsEndpoint = config.RightsEndpoint

	return impl
}

func (si *tracksysEnrichStepImpl) Name( ) string {
	return "Tracksys enrich"
}

func (si *tracksysEnrichStepImpl) Process(message *awssqs.Message, _ interface{}) (bool, interface{}, error) {

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

			// we have determined that we do not want to enrich certain class of item
			shouldEnrich := si.enrichableItem(message)
			if shouldEnrich == true {
				log.Printf("INFO: located id %s in tracksys cache, getting details", id)
				trackSysDetails, err := TracksysIdCache.Lookup(id)
				if err != nil {
					return false, nil, err
				}
				err = si.applyEnrichment(trackSysDetails, message)
				if err != nil {
					return false, nil, err
				}

				// we found the item in tracksys
				return true, trackSysDetails, nil
			} else {
				log.Printf("INFO: id %s is a special item, ignoring it", id)
			}
		}
	} else {
		log.Printf("ERROR: no identifier attribute located for document, no enrichment possible")
		return false, nil, errorNoIdentifier
	}

	return false, nil, nil
}

// there are certain classes of item that should not be enriched, not sure why but at the moment tracksys times
// out when we request them.
func (si *tracksysEnrichStepImpl) enrichableItem(message *awssqs.Message) bool {

	// search for the "serials" facade field
	facetTag := ConstructFieldTagPair("pool_f_stored", "serials")
	if strings.Contains(string(message.Payload), facetTag) {
		log.Printf("INFO: found %s in payload", facetTag)
		return false
	}

	return true
}

func (si *tracksysEnrichStepImpl) applyEnrichment(tracksysDetails *TrackSysItemDetails, message *awssqs.Message) error {

	// extract the information from the tracksys structure
	format_facets, _ := si.extractFormatFacets(tracksysDetails)
	feature_facets, _ := si.extractFeatureFacets(tracksysDetails)
	source_facets, _ := si.extractSourceFacets(tracksysDetails)
	marc_display_facets := []string{"true"}

	additional_collection_facets, _ := si.extractAdditionalCollectionFacets(tracksysDetails)
	alternate_id_facets, _ := si.extractAlternateIdFacets(tracksysDetails)
	individual_call_number_display, _ := si.extractCallNumbers(tracksysDetails)
	//iiif_presentation_metadata_display, err := si.extractIIIFManifest( tracksysDetails )
	//if err != nil {
	//   return err
	//}
	thumbnail_url_display, _ := si.extractThumbnailUrlDisplay(tracksysDetails)
	rights_wrapper_url_display, _ := si.extractRightsWrapperUrlDisplay(tracksysDetails)
	rights_wrapper_display, _ := si.extractRightsWrapperDisplay(tracksysDetails)
	pdf_url_display, _ := si.extractPdfUrlDisplay(tracksysDetails)
	policy_facets, err := si.extractPolicyFacets(tracksysDetails)
	if err != nil {
		return err
	}
	despined_barcodes_display, _ := si.extractDespinedBarcodesDisplay(tracksysDetails)

	// build our additional tag data
	var additionalTags strings.Builder

	additionalTags.WriteString(ConstructFieldTagSet("format_f_stored", XmlEncodeValues(format_facets)))
	additionalTags.WriteString(ConstructFieldTagSet("feature_f_stored", XmlEncodeValues(feature_facets)))
	additionalTags.WriteString(ConstructFieldTagSet("source_f_stored", XmlEncodeValues(source_facets)))

	additionalTags.WriteString(ConstructFieldTagSet("marc_display_f_stored", XmlEncodeValues(marc_display_facets)))
	additionalTags.WriteString(ConstructFieldTagSet("additional_collection_f_stored", XmlEncodeValues(additional_collection_facets)))
	additionalTags.WriteString(ConstructFieldTagSet("alternate_id_f_stored", XmlEncodeValues(alternate_id_facets)))
	additionalTags.WriteString(ConstructFieldTagSet("individual_call_number_a", XmlEncodeValues(individual_call_number_display)))
	additionalTags.WriteString(ConstructFieldTagSet("thumbnail_url_a", XmlEncodeValues(thumbnail_url_display)))
	additionalTags.WriteString(ConstructFieldTagSet("rights_wrapper_url_a", XmlEncodeValues(rights_wrapper_url_display)))
	additionalTags.WriteString(ConstructFieldTagSet("rights_wrapper_a", XmlEncodeValues(rights_wrapper_display)))
	additionalTags.WriteString(ConstructFieldTagSet("pdf_url_a", XmlEncodeValues(pdf_url_display)))
	additionalTags.WriteString(ConstructFieldTagSet("policy_f_stored", XmlEncodeValues(policy_facets)))
	additionalTags.WriteString(ConstructFieldTagSet("despined_barcodes_a", XmlEncodeValues(despined_barcodes_display)))

	// a special case
	//buf := ConstructFieldTagSet( "iiif_presentation_metadata_a", XmlEncodeValues( iiif_presentation_metadata_display ) )
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

func (si *tracksysEnrichStepImpl) extractFormatFacets(tracksysDetails *TrackSysItemDetails) ([]string, error) {
	res := []string{"Online"}
	return res, nil
}

func (si *tracksysEnrichStepImpl) extractFeatureFacets(tracksysDetails *TrackSysItemDetails) ([]string, error) {
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

func (si *tracksysEnrichStepImpl) extractSourceFacets(tracksysDetails *TrackSysItemDetails) ([]string, error) {
	res := []string{"UVA Library Digital Repository"}
	return res, nil
}

func (si *tracksysEnrichStepImpl) extractAdditionalCollectionFacets(tracksysDetails *TrackSysItemDetails) ([]string, error) {
	res := make([]string, 0, 1)
	if len(tracksysDetails.Collection) != 0 {
		res = append(res, tracksysDetails.Collection)
	}
	return res, nil
}

func (si *tracksysEnrichStepImpl) extractAlternateIdFacets(tracksysDetails *TrackSysItemDetails) ([]string, error) {
	res := make([]string, 0, 10)
	for _, i := range tracksysDetails.Items {
		if len(i.Pid) != 0 {
			res = append(res, i.Pid)
		}
	}
	return res, nil
}

func (si *tracksysEnrichStepImpl) extractCallNumbers(tracksysDetails *TrackSysItemDetails) ([]string, error) {
	res := make([]string, 0, 10)
	for _, i := range tracksysDetails.Items {
		if len(i.CallNumber) != 0 {
			res = append(res, i.CallNumber)
		}
	}
	return res, nil
}

//func (e *tracksysEnrichStepImpl) extractIIIFManifest(tracksysDetails *TrackSysItemDetails) ([]string, error) {
//
//	urls := make([]string, 0, 10)
//	for _, i := range tracksysDetails.Items {
//		if len(i.BackendIIIFManifestUrl) != 0 {
//			urls = append(urls, i.BackendIIIFManifestUrl)
//		}
//	}
//
//	res := make([]string, 0, 10)
//	for _, i := range urls {
//		body, err := httpGet(i, e.httpClient)
//		if err == nil {
//			res = append(res, string(body))
//		} else {
//			log.Printf("ERROR: endpoint %s returns %s", i, err)
//			return nil, err
//		}
//	}
//
//	return res, nil
//}

func (si *tracksysEnrichStepImpl) extractThumbnailUrlDisplay(tracksysDetails *TrackSysItemDetails) ([]string, error) {
	res := make([]string, 0, 10)
	for _, i := range tracksysDetails.Items {
		if len(i.ThumbnailUrl) != 0 {
			res = append(res, i.ThumbnailUrl)
		}
	}
	return res, nil
}

func (si *tracksysEnrichStepImpl) extractRightsWrapperUrlDisplay(tracksysDetails *TrackSysItemDetails) ([]string, error) {
	res := make([]string, 0, 10)
	for _, i := range tracksysDetails.Items {
		if len(i.RightsWrapperUrl) != 0 {
			res = append(res, i.RightsWrapperUrl)
		}
	}
	return res, nil
}

func (si *tracksysEnrichStepImpl) extractRightsWrapperDisplay(tracksysDetails *TrackSysItemDetails) ([]string, error) {
	res := make([]string, 0, 10)
	for _, i := range tracksysDetails.Items {
		if len(i.RightsWrapperText) != 0 {
			res = append(res, i.RightsWrapperText)
		}
	}
	return res, nil
}

func (si *tracksysEnrichStepImpl) extractPdfUrlDisplay(tracksysDetails *TrackSysItemDetails) ([]string, error) {
	res := make([]string, 0, 1)
	if len(tracksysDetails.PdfServiceRoot) != 0 {
		res = append(res, tracksysDetails.PdfServiceRoot)
	}
	return res, nil
}

func (si *tracksysEnrichStepImpl) extractPolicyFacets(tracksysDetails *TrackSysItemDetails) ([]string, error) {

	res := make([]string, 0, 1)
	for _, i := range tracksysDetails.Items {
		if len(i.Pid) != 0 {
			url := fmt.Sprintf("%s/%s", si.RightsEndpoint, i.Pid)
			body, err := httpGet(url, si.httpClient)
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

func (si *tracksysEnrichStepImpl) extractDespinedBarcodesDisplay(tracksysDetails *TrackSysItemDetails) ([]string, error) {

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

//
// end of file
//
