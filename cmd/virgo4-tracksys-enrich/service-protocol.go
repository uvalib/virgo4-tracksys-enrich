package main

import (
	"encoding/json"
	"log"
)

func (cl *cacheLoaderImpl) protocolGetKnownIds(url string) (*TracksysKnown, error) {

	payload, err := httpGet(url, cl.httpClient)
	if err != nil {
		return nil, err
	}

	tk := TracksysKnown{}
	err = json.Unmarshal(payload, &tk)
	if err != nil {
		log.Printf("ERROR: json unmarshal of TracksysKnown: %s", err)
		return nil, err
	}

	log.Printf("INFO: received %d known items", len(tk.Items))
	return &tk, nil
}

func (cl *cacheLoaderImpl) protocolGetSirsiDetails(url string) (*TracksysSirsiItem, error) {

	payload, err := httpGet(url, cl.httpClient)
	if err != nil {
		return nil, err
	}

	tsSirsi := TracksysSirsiItem{}
	err = json.Unmarshal(payload, &tsSirsi)
	if err != nil {
		log.Printf("ERROR: json unmarshal of TracksysSirsiItem: %s", err)
		return nil, err
	}
	return &tsSirsi, nil
}

func (cl *cacheLoaderImpl) protocolGetPidDetails(url string) (*TracksysPart, error) {

	payload, err := httpGet(url, cl.httpClient)
	if err != nil {
		return nil, err
	}

	tsPid := TracksysPart{}
	err = json.Unmarshal(payload, &tsPid)
	if err != nil {
		log.Printf("ERROR: json unmarshal of TracksysPart: %s", err)
		return nil, err
	}
	return &tsPid, nil
}

//
// end of file
//
