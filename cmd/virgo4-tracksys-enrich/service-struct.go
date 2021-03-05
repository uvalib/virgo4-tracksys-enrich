package main

// TracksysSirsiItem - a "Sirsi" item from tracksys, can contain multiple parts
type TracksysSirsiItem struct {
	SirsiId        string `json:"sirsiId"`
	PdfServiceRoot string `json:"pdfServiceRoot"`
	Collection     string `json:"collection"`
	Items          []TracksysPart
}

// TracksysPart - a "part" item from tracksys, represents a digital item/asset
type TracksysPart struct {
	Pid                    string   `json:"pid"`
	CallNumber             string   `json:"callNumber"`
	Barcode                string   `json:"barcode"`
	RsURI                  string   `json:"rsURI"`
	RsUses                 []string `json:"rsUses"`
	RightsWrapperUrl       string   `json:"rightsWrapperUrl"`
	RightsWrapperText      string   `json:"rightsWrapperText"`
	BackendIIIFManifestUrl string   `json:"backendIIIFManifestUrl"`
	ThumbnailUrl           string   `json:"thumbnailUrl"`
	PdfServiceRoot         string   `json:"pdfServiceRoot,omitempty"`

	// a special field we add that does not appear in the response json
	OcrCandidate bool `json:"-"`
}

// TracksysPidItem - a pid item from tracksys
type TracksysPidItem struct {
	Id                 uint32 `json:"id"`
	Pid                string `json:"pid"`
	Type               string `json:"type"`
	Title              string `json:"title"`
	AvailabilityPolicy string `json:"availability_policy"`
	OcrHint            string `json:"ocr_hint"`
	OcrCandidate       bool   `json:"ocr_candidate"`
	OcrLanguageHint    string `json:"ocr_language_hint"`
}

// TracksysKnown - when we query tracksys about which items it has
type TracksysKnown struct {
	Items []string `json:"items"`
}

//
// end of file
//
