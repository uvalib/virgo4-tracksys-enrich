package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

var maxHttpRetries = 3
var retrySleepTime = 100 * time.Millisecond

func newHttpClient(maxConnections int, timeout int) *http.Client {

	return &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: maxConnections,
		},
		Timeout: time.Duration(timeout) * time.Second,
	}
}

func httpGet(url string, client *http.Client) ([]byte, error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("ERROR: GET failed with error (%s)", err)
		return nil, err
	}

	var response *http.Response
	count := 0
	for {
		start := time.Now()
		response, err = client.Do(req)
		duration := time.Since(start)
		log.Printf("INFO: GET %s (elapsed %d ms)", url, duration.Milliseconds())

		count++
		if err != nil {
			if canRetry(err) == false {
				log.Printf("ERROR: GET failed with error (%s)", err)
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

			defer response.Body.Close()

			if response.StatusCode != http.StatusOK {
				logLevel := "ERROR"
				// log not found as informational instead of as an error
				if response.StatusCode == http.StatusNotFound {
					logLevel = "INFO"
				}
				log.Printf("%s: GET failed with status %d", logLevel, response.StatusCode)

				body, _ := ioutil.ReadAll(response.Body)

				return body, fmt.Errorf("request returns HTTP %d", response.StatusCode)
			} else {
				body, err := ioutil.ReadAll(response.Body)
				if err != nil {
					return nil, err
				}

				//log.Printf( body )
				return body, nil
			}
		}
	}
}

// examines the error and decides if it can be retried
func canRetry(err error) bool {

	if strings.Contains(err.Error(), "operation timed out") == true {
		return true
	}

	if strings.Contains(err.Error(), "Client.Timeout exceeded") == true {
		return true
	}

	if strings.Contains(err.Error(), "write: broken pipe") == true {
		return true
	}

	if strings.Contains(err.Error(), "no such host") == true {
		return true
	}

	if strings.Contains(err.Error(), "network is down") == true {
		return true
	}

	return false
}

//
// end of file
//
