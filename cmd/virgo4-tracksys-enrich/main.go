package main

import (
	"log"
	"os"
	"time"

	"github.com/uvalib/virgo4-sqs-sdk/awssqs"
)

//
// main entry point
//
func main() {

	log.Printf("===> %s service staring up (version: %s) <===", os.Args[ 0 ], Version( ) )

	// Get config params and use them to init service context. Any issues are fatal
	cfg := LoadConfiguration()

	// load our AWS_SQS helper object
	aws, err := awssqs.NewAwsSqs( awssqs.AwsSqsConfig{ } )
	if err != nil {
		log.Fatal( err )
	}

	// get the queue handles from the queue name
	inQueueHandle, err := aws.QueueHandle( cfg.InQueueName )
	if err != nil {
		log.Fatal( err )
	}

	outQueue1Handle, err := aws.QueueHandle( cfg.OutQueue1Name )
	if err != nil {
		log.Fatal( err )
	}

	outQueue2Handle, err := aws.QueueHandle( cfg.OutQueue2Name )
	if err != nil {
		log.Fatal( err )
	}

	// create the record channel
	inboundMessageChan := make( chan awssqs.Message, cfg.WorkerQueueSize )

	// start workers here
	for w := 1; w <= cfg.Workers; w++ {
		go worker( w, cfg, aws, inboundMessageChan, inQueueHandle, outQueue1Handle, outQueue2Handle )
	}

    for {

		// wait for a batch of messages
		messages, err := aws.BatchMessageGet(inQueueHandle, awssqs.MAX_SQS_BLOCK_COUNT, time.Duration(cfg.PollTimeOut)*time.Second)
		if err != nil {
			log.Fatal(err)
		}

		// did we receive any?
		sz := len(messages)
		if sz != 0 {

			//log.Printf( "Received %d messages", sz )

			for _, m := range messages {
				inboundMessageChan <- m
			}

		} else {
			log.Printf("No messages available")
		}
	}
}

//
// end of file
//