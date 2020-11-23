package main

import (
	"bytes"
	"fmt"
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
	OembedUrl   string
}

// this is our actual implementation
type metadataCacheStepImpl struct {
	config  *ServiceConfig     // the service configuration
	s3proxy *S3Proxy           // our S3 abstraction
	tmpl    *template.Template // our pre-rendered template
}

// the field name in the SolrDoc
var metadataCacheFieldName = "digital_content_service_url_e_stored"

// NewMetaDataCacheStep - the factory
func NewMetaDataCacheStep(config *ServiceConfig) PipelineStep {

	// mock implementation here if necessary

	impl := &metadataCacheStepImpl{}
	impl.config = config
	impl.s3proxy = NewS3Proxy(config)
	impl.tmpl = template.Must(template.ParseFiles("templates/cache-entry.json"))
	return impl
}

func (si *metadataCacheStepImpl) Name() string {
	return "Metadata cache"
}

func (si *metadataCacheStepImpl) Process(message *awssqs.Message, data interface{}) (bool, interface{}, error) {

	tracksysData, ok := data.(TracksysSirsiItem)
	if ok == false {
		log.Printf("ERROR: failed to type assert into known payload")
		return false, data, ErrorTypeAssertion
	}

	err := si.createMetadataCache(tracksysData, message)
	if err != nil {
		return false, data, err
	}

	// if we were successful creating the metadata cache, include it's url in the SolrDoc

	current := string(message.Payload)
	metadataUrl := fmt.Sprintf("%s/%s/%s",
		si.config.DigitalContentCacheRoot,
		si.config.DigitalContentCacheBucket,
		tracksysData.SirsiId)
	current = AppendXmlField(current, metadataCacheFieldName, metadataUrl)
	message.Payload = []byte(current)

	return true, data, nil
}

func (si *metadataCacheStepImpl) createMetadataCache(tracksysDetails TracksysSirsiItem, message *awssqs.Message) error {

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

func (si *metadataCacheStepImpl) createMetadataContent(tracksysDetails TracksysSirsiItem, message *awssqs.Message) (string, error) {

	// build the dataset for the template generation
	md, err := si.buildTemplateData(tracksysDetails)
	if err != nil {
		log.Printf("ERROR: unable to build cache metadata for %s: %s", tracksysDetails.SirsiId, err.Error())
		return "", err
	}

	// render the template
	var outBuffer bytes.Buffer
	err = si.tmpl.Execute(&outBuffer, md)
	if err != nil {
		log.Printf("ERROR: unable to render cache metadata for %s: %s", tracksysDetails.SirsiId, err.Error())
		return "", err
	}
	log.Printf("INFO: cache metadata generated for %s", tracksysDetails.SirsiId)
	//log.Printf( outBuffer.String() )

	return outBuffer.String(), nil
}

func (si *metadataCacheStepImpl) buildTemplateData(tracksysDetails TracksysSirsiItem) (MetadataCache, error) {

	md := MetadataCache{}
	parts := make([]MetadataPart, 0)
	md.Id = tracksysDetails.SirsiId
	for _, item := range tracksysDetails.Items {
		part := MetadataPart{}

		part.ManifestUrl = item.BackendIIIFManifestUrl
		part.Label = item.CallNumber
		part.Pid = item.Pid
		part.ThumbUrl = item.ThumbnailUrl
		part.PdfUrl = fmt.Sprintf("%s/%s", tracksysDetails.PdfServiceRoot, item.Pid)
		part.OembedUrl = fmt.Sprintf("%s/%s", si.config.OembedRoot, item.Pid)

		parts = append(parts, part)
	}
	md.Parts = parts
	return md, nil
}

//
// end of file
//
