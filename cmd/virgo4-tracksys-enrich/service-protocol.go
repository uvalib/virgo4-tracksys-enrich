package main

import (
	//"bytes"
	//"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
	//"regexp"
	//"strconv"
	//"github.com/antchfx/xmlquery"
)

var maxHttpRetries = 3
var retrySleepTime = 100 * time.Millisecond

//var documentAddFailed = fmt.Errorf( "SOLR add failed" )

func (cl *cacheLoaderImpl) protocolDirectory(url string) ([]string, error) {

	body, err := cl.httpGet(url)
	if err != nil {
		return nil, err
	}

	// split the body into a set of identifiers
	tokens := strings.Split(string(body), ",")

	//for _, v := range tokens {
	//	fmt.Printf("[%s]\n", v )
	//}

	log.Printf("Received directory of %d items", len(tokens))
	return tokens, nil
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
		response, err = cl.httpClient.Do(req)
		count++
		if err != nil {
			if cl.canRetry(err) == false {
				return nil, err
			}

			// break when tried too many times
			if count >= maxHttpRetries {
				return nil, err
			}

			log.Printf("POST failed with error, retrying (%s)", err)

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

//func ( s * solrImpl ) processResponsePayload( body []byte ) ( int, uint, error ) {
//
//	// generate a query structure from the body
//	doc, err := xmlquery.Parse( bytes.NewReader( body ) )
//	if err != nil {
//		return 0, 0, err
//	}
//
//	// attempt to extract the statusNode field
//	statusNode := xmlquery.FindOne( doc, "//response/lst[@name='responseHeader']/int[@name='status']")
//	if statusNode == nil {
//		return 0, 0, fmt.Errorf( "Cannot find status field in response payload (%s)", body )
//	}
//
//	// if it appears that we have an error
//	if statusNode.InnerText( ) != "0" {
//
//		// extract the status and attempt to find the error messageNode body
//		status, _ := strconv.Atoi( statusNode.InnerText( ) )
//
//		messageNode := xmlquery.FindOne( doc, "//response/lst[@name='error']/str[@name='msg']")
//		if messageNode != nil {
//
//			// if this is an error on a specific document, we can extract that information
//			re := regexp.MustCompile(`\[(\d+),\d+\]`)
//			match := re.FindStringSubmatch( messageNode.InnerText( ) )
//			if match != nil {
//				fmt.Printf( "%s", body )
//				docnum, _ := strconv.Atoi( match[ 1 ] )
//
//				// return index of failing item
//				return status, uint( docnum ) - 1, documentAddFailed
//			}
//		}
//		return status, 0, fmt.Errorf( "%s", body )
//	}
//
//	// all good
//    return 0, 0, nil
//}

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
