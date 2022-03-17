package main

import (
	"bytes"
	"fmt"
	"github.com/uvalib/virgo4-sqs-sdk/awssqs"
	"log"
	"text/template"
)

// MetadataCache - the structure we will template for the metadata cache entry
type MetadataCache struct {
	Id    string
	Parts []MetadataPart
}

// MetadataPart - the structure we will template for the metadata cache entry
type MetadataPart struct {
	ManifestUrl string
	Label       string
	Pid         string
	ThumbUrl    string
	OcrUrl      string
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
	if config.Mode == "sirsi" {
		impl.tmpl = template.Must(template.ParseFiles("templates/multi-pid-cache-entry.json"))
	} else {
		impl.tmpl = template.Must(template.ParseFiles("templates/single-pid-cache-entry.json"))
	}
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

	key, err := si.createMetadataCache(tracksysData, message)
	if err != nil {
		return false, data, err
	}

	// if we were successful creating the metadata cache, include it's url in the SolrDoc

	current := string(message.Payload)
	metadataUrl := fmt.Sprintf("%s/%s/%s",
		si.config.DigitalContentCacheRoot,
		si.config.DigitalContentCacheBucket,
		key)
	//log.Printf("METADATA URL: %s", metadataUrl)
	current = AppendXmlField(current, metadataCacheFieldName, metadataUrl)
	message.Payload = []byte(current)

	return true, data, nil
}

func (si *metadataCacheStepImpl) createMetadataCache(tracksysDetails TracksysSirsiItem, message *awssqs.Message) (string, error) {

	var err error
	var metadata string
	var key string
	if si.config.Mode == "sirsi" {
		key = tracksysDetails.SirsiId
		metadata, err = si.createMultiPidMetadataContent(tracksysDetails)
	} else {
		key = normalizeId(tracksysDetails.Items[0].Pid)
		metadata, err = si.createSinglePidMetadataContent(tracksysDetails)
	}

	if err != nil {
		return "", err
	}

	err = si.s3proxy.WriteToCache(key, metadata)
	if err != nil {
		return "", err
	}
	return key, nil
}

func (si *metadataCacheStepImpl) createMultiPidMetadataContent(tracksysDetails TracksysSirsiItem) (string, error) {

	// build the dataset for the template generation
	td := si.buildMultiPidTemplateData(tracksysDetails)

	// render the template
	var outBuffer bytes.Buffer
	err := si.tmpl.Execute(&outBuffer, td)
	if err != nil {
		log.Printf("ERROR: unable to render cache metadata for %s: %s", td.Id, err.Error())
		return "", err
	}
	log.Printf("INFO: cache metadata generated for %s", td.Id)
	//log.Printf(outBuffer.String())

	return outBuffer.String(), nil
}

func (si *metadataCacheStepImpl) createSinglePidMetadataContent(tracksysDetails TracksysSirsiItem) (string, error) {

	// build the dataset for the template generation
	td := si.buildSinglePidTemplateData(tracksysDetails)

	// render the template
	var outBuffer bytes.Buffer
	err := si.tmpl.Execute(&outBuffer, td)
	if err != nil {
		log.Printf("ERROR: unable to render cache metadata for %s: %s", td.Pid, err.Error())
		return "", err
	}
	log.Printf("INFO: cache metadata generated for %s", td.Pid)
	//log.Printf(outBuffer.String())

	return outBuffer.String(), nil
}

func (si *metadataCacheStepImpl) buildMultiPidTemplateData(tracksysDetails TracksysSirsiItem) MetadataCache {

	mc := MetadataCache{}
	parts := make([]MetadataPart, 0)
	mc.Id = tracksysDetails.SirsiId
	for _, item := range tracksysDetails.Items {
		part := MetadataPart{}

		part.ManifestUrl = item.BackendIIIFManifestUrl
		part.Label = item.CallNumber
		part.Pid = item.Pid
		part.ThumbUrl = item.ThumbnailUrl
		if item.OcrCandidate == true {
			log.Printf("INFO: PID %s is an OCR candidate", item.Pid)
			part.OcrUrl = fmt.Sprintf("%s/%s", si.config.OcrServiceRoot, item.Pid)
		}
		part.PdfUrl = fmt.Sprintf("%s/%s", tracksysDetails.PdfServiceRoot, item.Pid)
		part.OembedUrl = fmt.Sprintf("%s/%s", si.config.OembedRoot, item.Pid)

		parts = append(parts, part)
	}
	mc.Parts = parts
	return mc
}

func (si *metadataCacheStepImpl) buildSinglePidTemplateData(tracksysDetails TracksysSirsiItem) MetadataPart {

	mp := MetadataPart{}
	mp.Pid = tracksysDetails.Items[0].Pid
	mp.ManifestUrl = tracksysDetails.Items[0].BackendIIIFManifestUrl
	mp.Label = tracksysDetails.Items[0].CallNumber
	mp.Pid = tracksysDetails.Items[0].Pid
	mp.ThumbUrl = tracksysDetails.Items[0].ThumbnailUrl
	mp.PdfUrl = fmt.Sprintf("%s/%s", tracksysDetails.Items[0].PdfServiceRoot, mp.Pid)
	mp.OembedUrl = fmt.Sprintf("%s/%s", si.config.OembedRoot, mp.Pid)
	return mp
}

//
// end of file
//
