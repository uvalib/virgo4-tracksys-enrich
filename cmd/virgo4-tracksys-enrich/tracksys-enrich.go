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

// FIXME
var rightsServiceEndpoint = "http://rightsws.lib.virginia.edu:8089"

//
// Much of this code is based on the existing SOLRMark plugin to handle tracksys decoration. See more details:
// https://github.com/uvalib/utilities/blob/master/bib/bin/solrmarc3/index_java/src/DlMixin.java
// https://github.com/uvalib/utilities/blob/master/bib/bin/solrmarc3/dl_augment.properties
//

func applyEnrichment(message * awssqs.Message, tracksysDetails * TrackSysItemDetails ) error {

   // extract the information from the tracksys structure
   format_facets                      := extractFormatFacets( tracksysDetails )
   feature_facets                     := extractFeatureFacets( tracksysDetails )
   source_facets                      := extractSourceFacets( tracksysDetails )
   marc_display_facets                := []string{"true"}

   additional_collection_facets       := extractAdditionalCollectionFacets( tracksysDetails )
   alternate_id_facets                := extractAlternateIdFacets( tracksysDetails )
   individual_call_number_display     := extractCallNumbers( tracksysDetails )
   //iiif_presentation_metadata_display := extractIIIFManifest( tracksysDetails )
   thumbnail_url_display              := extractThumbnailUrlDisplay( tracksysDetails )
   rights_wrapper_url_display         := extractRightsWrapperUrlDisplay( tracksysDetails )
   rights_wrapper_display             := extractRightsWrapperDisplay( tracksysDetails )
   pdf_url_display                    := extractPdfUrlDisplay( tracksysDetails )
   policy_facets                      := extractPolicyFacets( tracksysDetails )
   despined_barcodes_display          := extractDespinedBarcodesDisplay( tracksysDetails )

   // build our additional tag data
   var additionalTags strings.Builder

   additionalTags.WriteString( makeFieldTagSet( "format_f_stored", format_facets ) )
   additionalTags.WriteString( makeFieldTagSet( "feature_f_stored", feature_facets ) )
   additionalTags.WriteString( makeFieldTagSet( "source_f_stored", source_facets ) )

   additionalTags.WriteString( makeFieldTagSet( "marc_display_f_stored", marc_display_facets ) )
   additionalTags.WriteString( makeFieldTagSet( "additional_collection_f_stored", additional_collection_facets ) )
   additionalTags.WriteString( makeFieldTagSet( "alternate_id_f_stored", alternate_id_facets ) )
   additionalTags.WriteString( makeFieldTagSet( "individual_call_number_a", individual_call_number_display ) )
   //additionalTags.WriteString( makeFieldTagSet( "iiif_presentation_metadata_a", iiif_presentation_metadata_display ) )
   additionalTags.WriteString( makeFieldTagSet( "thumbnail_url_a", thumbnail_url_display ) )
   additionalTags.WriteString( makeFieldTagSet( "rights_wrapper_url_a", rights_wrapper_url_display ) )
   additionalTags.WriteString( makeFieldTagSet( "rights_wrapper_a", rights_wrapper_display ) )
   additionalTags.WriteString( makeFieldTagSet( "pdf_url_a", pdf_url_display ) )
   additionalTags.WriteString( makeFieldTagSet( "policy_f_stored", policy_facets ) )
   additionalTags.WriteString( makeFieldTagSet( "despined_barcodes_a", despined_barcodes_display ) )

   // tack it on the end of the document
   docEndTag := "</doc>"
   //log.Printf( "Enrich with [%s]", additionalTags.String() )
   additionalTags.WriteString( docEndTag )
   current := string( message.Payload )
   current = strings.Replace( current, docEndTag, additionalTags.String(), 1 )
   message.Payload = []byte( current )
   return nil
}

func extractFormatFacets( tracksysDetails * TrackSysItemDetails ) []string {
   res := []string{ "Online" }
   return res
}

func extractFeatureFacets ( tracksysDetails * TrackSysItemDetails ) []string {
   res := make( []string, 0, 5 )
   res = append(res, "availability" )
   res = append(res, "iiif" )
   res = append(res, "dl_metadata" )
   res = append(res, "rights_wrapper" )
   if len( tracksysDetails.PdfServiceRoot ) != 0 {
      res = append(res, "pdf_service" )
   }
   return res
}

func extractSourceFacets( tracksysDetails * TrackSysItemDetails ) []string {
   res := []string{ "UVA Library Digital Repository" }
   return res
}

func extractAdditionalCollectionFacets( tracksysDetails * TrackSysItemDetails ) []string {
   res := make( []string, 0, 1 )
   if len( tracksysDetails.Collection ) != 0 {
      res = append(res, tracksysDetails.Collection )
   }
   return res
}

func extractAlternateIdFacets( tracksysDetails * TrackSysItemDetails ) []string {
   res := make( []string, 0, 10 )
   for _, i := range tracksysDetails.Items {
      if len( i.Pid ) != 0 {
         res = append(res, i.Pid )
      }
   }
   return res
}

func extractCallNumbers( tracksysDetails * TrackSysItemDetails ) []string {
   res := make( []string, 0, 10 )
   for _, i := range tracksysDetails.Items {
      if len( i.CallNumber ) != 0 {
         res = append(res, i.CallNumber )
      }
   }
   return res
}

func extractIIIFManifest( tracksysDetails * TrackSysItemDetails ) []string {

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
         log.Printf("ERROR: endpoint %s returns %s (ignoring)", i, err )
      }
   }

   return res
}

func extractThumbnailUrlDisplay( tracksysDetails * TrackSysItemDetails ) []string {
   res := make( []string, 0, 10 )
   for _, i := range tracksysDetails.Items {
      if len( i.ThumbnailUrl ) != 0 {
         res = append(res, i.ThumbnailUrl )
      }
   }
   return res
}

func extractRightsWrapperUrlDisplay( tracksysDetails * TrackSysItemDetails ) []string {
   res := make( []string, 0, 10 )
   for _, i := range tracksysDetails.Items {
      if len( i.RightsWrapperUrl ) != 0 {
         res = append(res, i.RightsWrapperUrl )
      }
   }
   return res
}

func extractRightsWrapperDisplay( tracksysDetails * TrackSysItemDetails ) []string {
   res := make( []string, 0, 10 )
   for _, i := range tracksysDetails.Items {
      if len( i.RightsWrapperText ) != 0 {
         res = append(res, i.RightsWrapperText )
      }
   }
   return res
}

func extractPdfUrlDisplay( tracksysDetails * TrackSysItemDetails ) []string {
   res := make( []string, 0, 1 )
   if len( tracksysDetails.PdfServiceRoot ) != 0 {
      res = append(res, tracksysDetails.PdfServiceRoot )
   }
   return res
}

func extractPolicyFacets( tracksysDetails * TrackSysItemDetails ) []string {

   res := make( []string, 0, 1 )
   for _, i := range tracksysDetails.Items {
      if len( i.Pid ) != 0 {
         url := fmt.Sprintf( "%s/%s", rightsServiceEndpoint, i.Pid )
         httpClient := newHttpClient( externalServiceTimeoutInSeconds )
         body, err := httpGet( url, httpClient )
         if err == nil {
            if string( body ) != "public" {
               res = append(res, string( body ) )
            }
            break
         } else {
            log.Printf("ERROR: endpoint %s returns %s (ignoring)", url, err )
         }
      }
   }

   return res
}

func extractDespinedBarcodesDisplay( tracksysDetails * TrackSysItemDetails ) []string {

   res := make( []string, 0, 10 )
   if tracksysDetails.Collection == "Gannon Collection" {
      for _, i := range tracksysDetails.Items {
         if len( i.Barcode ) != 0 {
            res = append(res, i.Barcode )
         }
      }
   }
   return res
}

func makeFieldTagSet( name string, values []string ) string {
   var res strings.Builder
   for _, v := range values {
      res.WriteString( makeFieldTagPair( name, v))
   }
   return res.String()

}

func makeFieldTagPair( name string, value string ) string {
   return fmt.Sprintf( "<field name=\"%s\">%s</field>", name, xmlEscapeText( value ) )
}

func xmlEscapeText( value string ) string {
   var escaped bytes.Buffer
   _ = xml.EscapeText( &escaped, []byte( value ) )
   return escaped.String( )
}

//
// end of file
//
