package main

type TrackSysItemDetails struct {
	SirsiId        string `json:"sirsiId"`
	PdfServiceRoot string `json:"pdfServiceRoot"`
	Collection     string `json:"collection"`
	Items          []TrackSysItem
}

type TrackSysItem struct {
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

//
// end of file
//
