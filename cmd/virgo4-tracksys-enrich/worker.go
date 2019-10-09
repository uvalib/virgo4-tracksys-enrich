package main

import (
	"fmt"
	"log"
	"time"

	"github.com/uvalib/virgo4-sqs-sdk/awssqs"
)

// time to wait for inbound messages before doing something else
var waitTimeout = 5 * time.Second

var errorNoIdentifier = fmt.Errorf("No identifier attribute located for document")

func worker(id int, config *ServiceConfig, aws awssqs.AWS_SQS, cache CacheLoader, inbound <-chan awssqs.Message, inQueue awssqs.QueueHandle, outQueue awssqs.QueueHandle) {

	// keep a list of the messages queued so we can delete them once they are sent to SOLR
	queued := make([]awssqs.Message, 0, awssqs.MAX_SQS_BLOCK_COUNT)
	var message awssqs.Message

	blocksize := uint(0)
	totalCount := uint(0)
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
			totalCount++

			// add it to the queued list
			queued = append(queued, message)
			if blocksize == awssqs.MAX_SQS_BLOCK_COUNT {
				err := processesInboundBlock(id, aws, cache, queued, inQueue, outQueue)
				if err != nil {
					log.Fatal(err)
				}

				// reset the counts
				blocksize = 0
				queued = queued[:0]
			}

			if totalCount%1000 == 0 {
				duration := time.Since(start)
				log.Printf("Worker %d: processed %d messages (%0.2f tps)", id, totalCount, float64(totalCount)/duration.Seconds())
			}

		} else {

			// we timed out, probably best to send anything pending
			if blocksize != 0 {
				err := processesInboundBlock(id, aws, cache, queued, inQueue, outQueue)
				if err != nil {
					log.Fatal(err)
				}

				duration := time.Since(start)
				log.Printf("Worker %d: processed %d messages (%0.2f tps)", id, totalCount, float64(totalCount)/duration.Seconds())

				// reset the counts
				blocksize = 0
				queued = queued[:0]
			}

			// reset the time
			start = time.Now()
		}
	}
}

func processesInboundBlock(id int, aws awssqs.AWS_SQS, cache CacheLoader, inboundMessages []awssqs.Message, inQueue awssqs.QueueHandle, outQueue awssqs.QueueHandle) error {

	// enrich as much as possible, in the event of an error, dont process the document further
	for ix, _ := range inboundMessages {
		err := enrichMessage(cache, inboundMessages[ix])
		if err != nil {
			log.Printf("WARNING: enrich failed for message %d (ignoring)", ix)
//			return err
		}
	}

	//
	// There is some magic here that I dont really like. The inboundMessages carry some hidden state information within them which
	// indicates that the message is an 'oversize' one so there are corresponding S3 assets that need to be lifecycle managed.
	//
	// In order to work around this, we create a new block of inboundMessages for the outbound journey
	//

	outboundMessages := make( []awssqs.Message, 0, awssqs.MAX_SQS_BLOCK_COUNT )

	for ix, _ := range inboundMessages {
		outboundMessages = append( outboundMessages, *contentClone( inboundMessages[ix] ) )
	}

	opStatus, err := aws.BatchMessagePut(outQueue, outboundMessages)
	if err != nil {
		if err != awssqs.OneOrMoreOperationsUnsuccessfulError {
			return err
		}
	}

	// we only delete the ones that completed successfully
	deleteMessages := make([]awssqs.Message, 0, awssqs.MAX_SQS_BLOCK_COUNT)

	// check the operation results
	for ix, op := range opStatus {
		if op == false {
			log.Printf("WARNING: message %d failed to send to queue", ix)
		} else {
			deleteMessages = append(deleteMessages, inboundMessages[ix])
		}
	}


	// delete the ones that succeeded
	opStatus, err = aws.BatchMessageDelete(inQueue, deleteMessages)
	if err != nil {
		if err != awssqs.OneOrMoreOperationsUnsuccessfulError {
			return err
		}
	}

	// check the operation results
	for ix, op := range opStatus {
		if op == false {
			log.Printf("WARNING: message %d failed to delete", ix)
		}
	}

	return err
}

func enrichMessage(cache CacheLoader, message awssqs.Message) error {

	id := getMessageAttribute(message, "id")
	if len(id) != 0 {
		found, err := cache.Contains(id)
		if err != nil {
			return err
		}

		// we have information about this item, pull it from Tracksys
		if found == true {
			log.Printf("INFO: located id %s in tracksys cache, getting details", id)
			tracksysDetails, err := cache.Lookup( id )
			if err != nil {
				return err
			}
			err = applyEnrichment( message, tracksysDetails )
			if err != nil {
				return err
			}
		}
	} else {
		log.Printf("ERROR: no identifier attribute located for document, no enrichment possible")
		return errorNoIdentifier
	}

	return nil
}

// we need to clone the inbound messages and use them for outbound messages. Because there is some hidden state information
// within a message. See the comment above
func contentClone( message awssqs.Message ) * awssqs.Message {

   newMessage := &awssqs.Message{}
   newMessage.Attribs = message.Attribs
   newMessage.Payload = message.Payload
   return newMessage
}

func getMessageAttribute(message awssqs.Message, attribute string) string {

	for _, a := range message.Attribs {
		if a.Name == attribute {
			return a.Value
		}
	}
	return ""
}

//
// end of file
//
