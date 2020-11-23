package main

// a "Sirsi" item from tracksys, can contain multiple parts
type TracksysSirsiItem struct {
	SirsiId        string `json:"sirsiId"`
	PdfServiceRoot string `json:"pdfServiceRoot"`
	Collection     string `json:"collection"`
	Items          []TracksysPart
}

// a "part" item from tracksys, represents a digital item/asset
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
}

// when we query tracksys about which items it has
type TracksysKnown struct {
	Items []string `json:"items"`
}

//
// end of file
//
