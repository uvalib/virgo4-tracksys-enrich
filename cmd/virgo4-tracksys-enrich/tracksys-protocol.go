package main

import (

	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

var maxHttpRetries = 3
var retrySleepTime = 100 * time.Millisecond

func (cl *cacheLoaderImpl) protocolDirectory(url string) ([]string, error) {

	body, err := cl.httpGet(url)
	if err != nil {
		return nil, err
	}

	// split the body into a set of identifiers
	tokens := strings.Split(string(body), ",")

	log.Printf("Received directory of %d items", len(tokens))
	return tokens, nil
}

func (cl *cacheLoaderImpl) protocolDetails(url string) ([]byte, error) {

	body, err := cl.httpGet(url)
	if err != nil {
		return nil, err
	}

	//fmt.Printf("BODY: %s\n", body )
	return body, err
}

func (cl *cacheLoaderImpl) httpGet(url string) ([]byte, error) {

	//fmt.Printf( "%s\n", s.url )

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	//req.Header.Set("Content-Type", "application/xml" )
	//req.Header.Set("Accept", "application/json" )

	var response *http.Response
	count := 0
	for {
		start := time.Now()
		response, err = cl.httpClient.Do(req)
		duration := time.Since(start)
		log.Printf("INFO: GET %s (elapsed %d ms)", url, duration.Milliseconds())

		count++
		if err != nil {
			if cl.canRetry(err) == false {
				return nil, err
			}

			// break when tried too many times
			if count >= maxHttpRetries {
				return nil, err
			}

			log.Printf("ERROR: GET failed with error, retrying (%s)", err)

			// sleep for a bit before retrying
			time.Sleep(retrySleepTime)
		} else {
			// success, break
			break
		}
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	//log.Printf( body )
	return body, nil
}

// examines the error and decides if if can be retried
func (cl *cacheLoaderImpl) canRetry(err error) bool {

	if strings.Contains(err.Error(), "operation timed out") == true {
		return true
	}

	if strings.Contains(err.Error(), "write: broken pipe") == true {
		return true
	}

	//if strings.Contains( err.Error( ), "network is down" ) == true {
	//	return true
	//}

	return false
}

//
// end of file
//
