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

// Enricher - the interface
type Enricher interface {
	Enrich(CacheLoader, *awssqs.Message) (bool, bool, error)
}

// this is our actual implementation
type enrichImpl struct {
	RightsEndpoint string       // the Rights URL
	httpClient     *http.Client // our http client connection
}

// NewEnricher - the factory
func NewEnricher(config *ServiceConfig) PipelineStep {

	// mock implementation here if necessary

	impl := &enrichImpl{}

	impl.httpClient = newHttpClient(2, config.ServiceTimeout)
	impl.RightsEndpoint = config.RightsEndpoint

	return impl
}

func (r *enrichImpl) Name( ) string {
	return "Tracksys enrich"
}

func (e *enrichImpl) Process(message *awssqs.Message) (bool, bool, error) {

	// passed back to caller in the event there are subsequent processing steps
	inTrackSys := false
	wasEnriched := false

	// extract the ID else we cannot do anything
	id, foundId := message.GetAttribute(awssqs.AttributeKeyRecordId)
	if foundId == true {
		var err error
		inTrackSys, err = TracksysIdCache.Contains(id)
		if err != nil {
			return inTrackSys, wasEnriched, err
		}

		// tracksys contains information about this item
		if inTrackSys == true {

			// we have determined that we do not want to enrich certain class of item
			shouldEnrich := e.enrichableItem(message)
			if shouldEnrich == true {
				log.Printf("INFO: located id %s in tracksys cache, getting details", id)
				trackSysDetails, err := TracksysIdCache.Lookup(id)
				if err != nil {
					return inTrackSys, wasEnriched, err
				}
				err = e.applyEnrichment(trackSysDetails, message)
				if err != nil {
					return inTrackSys, wasEnriched, err
				}
				// we did some sort of enrichment
				wasEnriched = true
			} else {
				log.Printf("INFO: id %s is a special item, ignoring it", id)
			}
		}
	} else {
		log.Printf("ERROR: no identifier attribute located for document, no enrichment possible")
		return inTrackSys, wasEnriched, errorNoIdentifier
	}

	return inTrackSys, wasEnriched, nil
}

// there are certain classes of item that should not be enriched, not sure why but at the moment tracksys times
// out when we request them.
func (e *enrichImpl) enrichableItem(message *awssqs.Message) bool {

	// search for the "serials" facade field
	facetTag := ConstructFieldTagPair("pool_f_stored", "serials")
	if strings.Contains(string(message.Payload), facetTag) {
		log.Printf("INFO: found %s in payload", facetTag)
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

//func (e *enrichImpl) extractIIIFManifest(tracksysDetails *TrackSysItemDetails) ([]string, error) {
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

//
// end of file
//
