package main

import (
	"bytes"
	"github.com/uvalib/virgo4-sqs-sdk/awssqs"
	"log"
	"text/template"
)

// the structure we will template for the metadata cache entry
type MetadataCache struct {
	Id    string
	Parts []MetadataPart
}

type MetadataPart struct {
	ManifestUrl string
	Label       string
	Pid         string
	ThumbUrl    string
	PdfUrl      string
	PdfStatus   string
}

// this is our actual implementation
type metadataCacheStepImpl struct {
	s3proxy *S3Proxy            // our S3 abstraction
	tmpl    *template.Template  // our pre-rendered template
}

// NewMetaDataCacheStep - the factory
func NewMetaDataCacheStep(config *ServiceConfig) PipelineStep {

	// mock implementation here if necessary

	impl := &metadataCacheStepImpl{}
	impl.s3proxy = NewS3Proxy(config)
	impl.tmpl = template.Must(template.ParseFiles("templates/cache-entry.json"))
	return impl
}

func (si *metadataCacheStepImpl) Name() string {
	return "Metadata cache"
}

func (si *metadataCacheStepImpl) Process(message *awssqs.Message, data interface{}) (bool, interface{}, error) {

	tracksysData, ok := data.(TrackSysItemDetails)
	if ok == false {
		log.Printf("ERROR: failed to type assert into known payload")
		return false, data, ErrorTypeAssertion
	}

	err := si.createMetadataCache(tracksysData, message)
	if err != nil {
		return false, data, err
	}

	return true, data, nil
}

func (si *metadataCacheStepImpl) createMetadataCache(tracksysDetails TrackSysItemDetails, message *awssqs.Message) error {

	metadata, err := si.createMetadataContent(tracksysDetails, message)
	if err != nil {
		return err
	}

	key := tracksysDetails.SirsiId
	err = si.s3proxy.WriteToCache(key, metadata)
	if err != nil {
		return err
	}
	return nil
}

func (si *metadataCacheStepImpl) createMetadataContent(tracksysDetails TrackSysItemDetails, message *awssqs.Message) (string, error) {

	// TEMP ONLY
	data := MetadataCache{}
	part := MetadataPart{}

	part.Label = "A B C"
	part.ManifestUrl = "https://iiifman.lib.virginia.edu/pid/uva-lib:1234"
	part.PdfStatus = "READY"
	part.PdfUrl = "https://pdfservice.lib.virginia.edu/pdf/uva-lib:1234"
	part.Pid = "uva-lib:1234"
    part.ThumbUrl = "https://iiif.lib.virginia.edu/iiif/uva-lib:1234/full/!125,200/0/default.jpg"

	data.Id = tracksysDetails.SirsiId
	data.Parts = []MetadataPart{ part }

	var outBuffer bytes.Buffer
	err := si.tmpl.Execute(&outBuffer, data)
	if err != nil {
		log.Printf("ERROR: unable to render cache metadata for %s: %s", data.Id, err.Error())
		return "", err
	}
	log.Printf("INFO: cache metadata generated for %s", data.Id)
	log.Printf( outBuffer.String() )

	return outBuffer.String(), nil
}

//
// end of file
//
