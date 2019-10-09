package main

import (
	"log"
	"strings"
)

func (cl *cacheLoaderImpl) protocolDirectory(url string) ([]string, error) {

	body, err := httpGet(url, cl.httpClient)
	if err != nil {
		return nil, err
	}

	// split the body into a set of identifiers
	tokens := strings.Split(string(body), ",")

	log.Printf("Received directory of %d items", len(tokens))
	return tokens, nil
}

func (cl *cacheLoaderImpl) protocolDetails(url string) ([]byte, error) {

	body, err := httpGet(url, cl.httpClient)
	if err != nil {
		return nil, err
	}

	//fmt.Printf("BODY: %s\n", body )
	return body, err
}


//
// end of file
//
