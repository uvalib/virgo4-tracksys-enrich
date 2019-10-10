package main

import (
   "bytes"
   "encoding/xml"
   //"bytes"
   "fmt"
   "log"
   "strings"

    "github.com/uvalib/virgo4-sqs-sdk/awssqs"
)

// just made this up...
var externalServiceTimeoutInSeconds = 15

// a SOLR limitation
//var maxSolrFieldSize = 32765

//
// Much of this code is based on the existing SOLRMark plugin to handle tracksys decoration. See more details:
// https://github.com/uvalib/utilities/blob/master/bib/bin/solrmarc3/index_java/src/DlMixin.java
// https://github.com/uvalib/utilities/blob/master/bib/bin/solrmarc3/dl_augment.properties
//

func applyEnrichment(config *ServiceConfig, tracksysDetails * TrackSysItemDetails, message * awssqs.Message ) error {

   // extract the information from the tracksys structure
   format_facets, _                      := extractFormatFacets( tracksysDetails )
   feature_facets, _                     := extractFeatureFacets( tracksysDetails )
   source_facets, _                      := extractSourceFacets( tracksysDetails )
   marc_display_facets                   := []string{"true"}

   additional_collection_facets, _       := extractAdditionalCollectionFacets( tracksysDetails )
   alternate_id_facets, _                := extractAlternateIdFacets( tracksysDetails )
   individual_call_number_display, _     := extractCallNumbers( tracksysDetails )
   //iiif_presentation_metadata_display, err := extractIIIFManifest( tracksysDetails )
   //if err != nil {
   //   return err
   //}
   thumbnail_url_display, _              := extractThumbnailUrlDisplay( tracksysDetails )
   rights_wrapper_url_display, _         := extractRightsWrapperUrlDisplay( tracksysDetails )
   rights_wrapper_display, _             := extractRightsWrapperDisplay( tracksysDetails )
   pdf_url_display, _                    := extractPdfUrlDisplay( tracksysDetails )
   policy_facets, err                    := extractPolicyFacets( config.RightsEndpoint, tracksysDetails )
   if err != nil {
      return err
   }
   despined_barcodes_display, _          := extractDespinedBarcodesDisplay( tracksysDetails )

   // build our additional tag data
   var additionalTags strings.Builder

   additionalTags.WriteString( makeFieldTagSet( "format_f_stored", xmlEncodeValues( format_facets ) ) )
   additionalTags.WriteString( makeFieldTagSet( "feature_f_stored", xmlEncodeValues( feature_facets ) ) )
   additionalTags.WriteString( makeFieldTagSet( "source_f_stored", xmlEncodeValues( source_facets ) ) )

   additionalTags.WriteString( makeFieldTagSet( "marc_display_f_stored", xmlEncodeValues( marc_display_facets ) ) )
   additionalTags.WriteString( makeFieldTagSet( "additional_collection_f_stored", xmlEncodeValues( additional_collection_facets ) ) )
   additionalTags.WriteString( makeFieldTagSet( "alternate_id_f_stored", xmlEncodeValues( alternate_id_facets ) ) )
   additionalTags.WriteString( makeFieldTagSet( "individual_call_number_a", xmlEncodeValues( individual_call_number_display ) ) )
   additionalTags.WriteString( makeFieldTagSet( "thumbnail_url_a", xmlEncodeValues( thumbnail_url_display ) ) )
   additionalTags.WriteString( makeFieldTagSet( "rights_wrapper_url_a", xmlEncodeValues( rights_wrapper_url_display ) ) )
   additionalTags.WriteString( makeFieldTagSet( "rights_wrapper_a", xmlEncodeValues( rights_wrapper_display ) ) )
   additionalTags.WriteString( makeFieldTagSet( "pdf_url_a", xmlEncodeValues( pdf_url_display ) ) )
   additionalTags.WriteString( makeFieldTagSet( "policy_f_stored", xmlEncodeValues( policy_facets ) ) )
   additionalTags.WriteString( makeFieldTagSet( "despined_barcodes_a", xmlEncodeValues( despined_barcodes_display ) ) )

   // a special case
   //buf := makeFieldTagSet( "iiif_presentation_metadata_a", xmlEncodeValues( iiif_presentation_metadata_display ) )
   //sz := len( buf )
   //if sz > maxSolrFieldSize {
   //   log.Printf("WARNING: iiif_presentation_metadata_a field exceeds maximum size, ignoring. size %d", sz )
   //} else {
   //   additionalTags.WriteString(buf)
   //}

   // tack it on the end of the document
   docEndTag := "</doc>"
   //log.Printf( "Enrich with [%s]", additionalTags.String() )
   additionalTags.WriteString( docEndTag )
   current := string( message.Payload )
   current = strings.Replace( current, docEndTag, additionalTags.String(), 1 )
   message.Payload = []byte( current )
   return nil
}

func extractFormatFacets( tracksysDetails * TrackSysItemDetails ) ( []string, error ) {
   res := []string{ "Online" }
   return res, nil
}

func extractFeatureFacets ( tracksysDetails * TrackSysItemDetails ) ( []string, error ) {
   res := make( []string, 0, 5 )
   res = append(res, "availability" )
   res = append(res, "iiif" )
   res = append(res, "dl_metadata" )
   res = append(res, "rights_wrapper" )
   if len( tracksysDetails.PdfServiceRoot ) != 0 {
      res = append(res, "pdf_service" )
   }
   return res, nil
}

func extractSourceFacets( tracksysDetails * TrackSysItemDetails ) ( []string, error ) {
   res := []string{ "UVA Library Digital Repository" }
   return res, nil
}

func extractAdditionalCollectionFacets( tracksysDetails * TrackSysItemDetails ) ( []string, error ) {
   res := make( []string, 0, 1 )
   if len( tracksysDetails.Collection ) != 0 {
      res = append(res, tracksysDetails.Collection )
   }
   return res, nil
}

func extractAlternateIdFacets( tracksysDetails * TrackSysItemDetails ) ( []string, error ) {
   res := make( []string, 0, 10 )
   for _, i := range tracksysDetails.Items {
      if len( i.Pid ) != 0 {
         res = append(res, i.Pid )
      }
   }
   return res, nil
}

func extractCallNumbers( tracksysDetails * TrackSysItemDetails ) ( []string, error ) {
   res := make( []string, 0, 10 )
   for _, i := range tracksysDetails.Items {
      if len( i.CallNumber ) != 0 {
         res = append(res, i.CallNumber )
      }
   }
   return res, nil
}

func extractIIIFManifest( tracksysDetails * TrackSysItemDetails ) ( []string, error ) {

   urls := make( []string, 0, 10 )
   for _, i := range tracksysDetails.Items {
      if len( i.BackendIIIFManifestUrl ) != 0 {
         urls = append(urls, i.BackendIIIFManifestUrl )
      }
   }

   res := make( []string, 0, 10 )
   httpClient := newHttpClient( externalServiceTimeoutInSeconds )
   for _, i := range urls {
      body, err := httpGet( i, httpClient )
      if err == nil {
         res = append(res, string( body ) )
      } else {
         log.Printf("ERROR: endpoint %s returns %s", i, err )
         return nil, err
      }
   }

   return res, nil
}

func extractThumbnailUrlDisplay( tracksysDetails * TrackSysItemDetails ) ( []string, error ) {
   res := make( []string, 0, 10 )
   for _, i := range tracksysDetails.Items {
      if len( i.ThumbnailUrl ) != 0 {
         res = append(res, i.ThumbnailUrl )
      }
   }
   return res, nil
}

func extractRightsWrapperUrlDisplay( tracksysDetails * TrackSysItemDetails ) ( []string, error ) {
   res := make( []string, 0, 10 )
   for _, i := range tracksysDetails.Items {
      if len( i.RightsWrapperUrl ) != 0 {
         res = append(res, i.RightsWrapperUrl )
      }
   }
   return res, nil
}

func extractRightsWrapperDisplay( tracksysDetails * TrackSysItemDetails ) ( []string, error ) {
   res := make( []string, 0, 10 )
   for _, i := range tracksysDetails.Items {
      if len( i.RightsWrapperText ) != 0 {
         res = append(res, i.RightsWrapperText )
      }
   }
   return res, nil
}

func extractPdfUrlDisplay( tracksysDetails * TrackSysItemDetails ) ( []string, error ) {
   res := make( []string, 0, 1 )
   if len( tracksysDetails.PdfServiceRoot ) != 0 {
      res = append(res, tracksysDetails.PdfServiceRoot )
   }
   return res, nil
}

func extractPolicyFacets( rightsUrl string, tracksysDetails * TrackSysItemDetails ) ( []string, error ) {

   res := make( []string, 0, 1 )
   for _, i := range tracksysDetails.Items {
      if len( i.Pid ) != 0 {
         url := fmt.Sprintf( "%s/%s", rightsUrl, i.Pid )
         httpClient := newHttpClient( externalServiceTimeoutInSeconds )
         body, err := httpGet( url, httpClient )
         if err == nil {
            if string( body ) != "public" {
               res = append(res, string( body ) )
            }
            break
         } else {
            log.Printf("ERROR: endpoint %s returns %s", url, err )
            return nil, err
         }
      }
   }

   return res, nil
}

func extractDespinedBarcodesDisplay( tracksysDetails * TrackSysItemDetails ) ( []string, error ) {

   res := make( []string, 0, 10 )
   if tracksysDetails.Collection == "Gannon Collection" {
      for _, i := range tracksysDetails.Items {
         if len( i.Barcode ) != 0 {
            res = append(res, i.Barcode )
         }
      }
   }
   return res, nil
}

func xmlEncodeValues( values []string ) []string {
   for ix, _ := range values {
      values[ix] = xmlEscape( values[ix])
   }

   return values
}

func makeFieldTagSet( name string, values []string ) string {
   var res strings.Builder
   for _, v := range values {
      res.WriteString( makeFieldTagPair( name, v))
   }
   return res.String()
}

func makeFieldTagPair( name string, value string ) string {
   return fmt.Sprintf( "<field name=\"%s\">%s</field>", name, value )
}

func xmlEscape( value string ) string {
   var escaped bytes.Buffer
   _ = xml.EscapeText( &escaped, []byte( value ) )
   return escaped.String( )
}

//
// end of file
//
