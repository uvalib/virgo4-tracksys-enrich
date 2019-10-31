package main

import (
	"log"
	"time"

	"github.com/uvalib/virgo4-sqs-sdk/awssqs"
)

// time to wait for inbound messages before doing something else
var waitTimeout = 5 * time.Second

var emptyOpList = make([]awssqs.OpStatus, 0)

func worker(id int, config *ServiceConfig, aws awssqs.AWS_SQS, cache CacheLoader, inbound <-chan awssqs.Message, inQueue awssqs.QueueHandle, outQueue awssqs.QueueHandle) {

	// we use this to enrich each message as appropriate
	enricher := NewEnricher(config)

	// keep a list of the messages queued so we can delete them once they are sent to SOLR
	queued := make([]awssqs.Message, 0, awssqs.MAX_SQS_BLOCK_COUNT)
	var message awssqs.Message

	blocksize := uint(0)
	count := uint(0)
	start := time.Now()

	for {

		arrived := false

		// process a message or wait...
		select {
		case message = <-inbound:
			arrived = true
			break
		case <-time.After(waitTimeout):
			break
		}

		// we have an inbound message to process
		if arrived == true {

			// update counts
			blocksize++
			count++

			// add it to the queued list
			queued = append(queued, message)
			if blocksize == awssqs.MAX_SQS_BLOCK_COUNT {
				_, err := processesInboundBlock(enricher, aws, cache, queued, inQueue, outQueue)
				if err != nil {
					if err != awssqs.OneOrMoreOperationsUnsuccessfulError {
						fatalIfError(err)
					}
				}

				// reset the counts
				blocksize = 0
				queued = queued[:0]
			}

			if count%1000 == 0 {
				duration := time.Since(start)
				log.Printf("Worker %d: processed %d messages (%0.2f tps)", id, count, float64(count)/duration.Seconds())
			}

		} else {

			// we timed out, probably best to send anything pending
			if blocksize != 0 {
				_, err := processesInboundBlock(enricher, aws, cache, queued, inQueue, outQueue)
				if err != nil {
					if err != awssqs.OneOrMoreOperationsUnsuccessfulError {
						fatalIfError(err)
					}
				}

				duration := time.Since(start)
				log.Printf("Worker %d: processed %d messages (%0.2f tps) (flushing)", id, count, float64(count)/duration.Seconds())

				// reset the counts
				blocksize = 0
				queued = queued[:0]
			}

			// reset the metrics values
			count = 0
			start = time.Now()
		}
	}
}

func processesInboundBlock(enricher Enricher, aws awssqs.AWS_SQS, cache CacheLoader, inboundMessages []awssqs.Message, inQueue awssqs.QueueHandle, outQueue awssqs.QueueHandle) ([]awssqs.OpStatus, error) {

	// keep a list of the ones that succeed/fail
	finalStatus := make([]awssqs.OpStatus, len(inboundMessages))
	enrichStatus := make([]awssqs.OpStatus, len(inboundMessages))

	//log.Printf("%d records to process", len(inboundMessages))

	// enrich as much as possible, in the event of an error, dont process the document further
	for ix, _ := range inboundMessages {
		err := enricher.Enrich(cache, &inboundMessages[ix])

		// for now, we still want to process records that failed enrichment
		enrichStatus[ix] = true

		if err != nil {
			id, found := inboundMessages[ix].GetAttribute(awssqs.AttributeKeyRecordId)
			if found == false {
				log.Printf("WARNING: enrich failed for message %d (%s)", ix, err)
			} else {
				log.Printf("WARNING: enrich failed for id %s (%s)", id, err)
			}
		}
	}

	//
	// There is some magic here that I dont really like. The inboundMessages carry some hidden state information within them which
	// indicates that the message is an 'oversize' one so there are corresponding S3 assets that need to be lifecycle managed.
	//
	// In order to work around this, we create a new block of inboundMessages for the outbound journey
	//

	outboundMessages := make([]awssqs.Message, 0, len(inboundMessages))

	for ix, _ := range inboundMessages {
		// as long as the enrichment succeeded...
		if enrichStatus[ix] == true {
			outboundMessages = append(outboundMessages, *inboundMessages[ix].ContentClone())
		}
	}

	//log.Printf("%d records to publish", len(outboundMessages))

	putStatus, err := aws.BatchMessagePut(outQueue, outboundMessages)
	if err != nil {
		if err != awssqs.OneOrMoreOperationsUnsuccessfulError {
			return emptyOpList, err
		}
	}

	// check the operation results
	for ix, op := range putStatus {
		if op == false {
			log.Printf("WARNING: message %d failed to send to queue", ix)
		}
	}

	// we need to construct an array of results based on the operations performed, enrich and a put to the queue
	enrichErrors := 0
	for ix, v := range enrichStatus {
		finalStatus[ix] = true
		if v == false {
			finalStatus[ix] = false
			enrichErrors++
		} else {
			if putStatus[ix-enrichErrors] == false {
				finalStatus[ix] = false
			}
		}
	}

	// we only delete the ones that completed successfully
	deleteMessages := make([]awssqs.Message, 0, len(outboundMessages))

	for ix, op := range finalStatus {
		if op == true {
			deleteMessages = append(deleteMessages, inboundMessages[ix])
		}
	}

	//log.Printf("%d records to delete", len(deleteMessages))

	// delete the ones that succeeded
	delStatus, err := aws.BatchMessageDelete(inQueue, deleteMessages)
	if err != nil {
		if err != awssqs.OneOrMoreOperationsUnsuccessfulError {
			return emptyOpList, err
		}
	}

	// we will ignore delete failures for now because they will be tried again when the message is next processed
	for ix, op := range delStatus {
		if op == false {
			log.Printf("WARNING: message %d failed to delete", ix)
		}
	}

	return finalStatus, err
}

//
// end of file
//
