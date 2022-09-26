package main

import (
	"log"
	"os"
	"time"

	"github.com/uvalib/virgo4-sqs-sdk/awssqs"
)

// main entry point
func main() {

	log.Printf("===> %s service staring up (version: %s) <===", os.Args[0], Version())

	// Get config params and use them to init service context. Any issues are fatal
	cfg := LoadConfiguration()

	// load our AWS_SQS helper object
	aws, err := awssqs.NewAwsSqs(awssqs.AwsSqsConfig{MessageBucketName: cfg.MessageBucketName})
	fatalIfError(err)

	// get the queue handles from the queue name
	inQueueHandle, err := aws.QueueHandle(cfg.InQueueName)
	fatalIfError(err)

	outQueueHandle, err := aws.QueueHandle(cfg.OutQueueName)
	fatalIfError(err)

	// load the Tracksis ID cache (so we only lookup items in tracksys that we know already exist)
	err = NewCacheLoader(cfg)
	fatalIfError(err)

	// create the record channel
	inboundMessageChan := make(chan awssqs.Message, cfg.WorkerQueueSize)

	// start workers here
	for w := 1; w <= cfg.Workers; w++ {
		go worker(w, cfg, aws, inboundMessageChan, inQueueHandle, outQueueHandle)
	}

	for {

		// wait for a batch of messages
		messages, err := aws.BatchMessageGet(inQueueHandle, awssqs.MAX_SQS_BLOCK_COUNT, time.Duration(cfg.PollTimeOut)*time.Second)
		if err != nil {
			log.Printf("ERROR: during message get (%s), sleeping and retrying", err.Error())

			// sleep for a while
			time.Sleep(1 * time.Second)

			// and try again
			continue
		}

		// did we receive any?
		sz := len(messages)
		if sz != 0 {

			//log.Printf( "Received %d messages", sz )

			for _, m := range messages {
				inboundMessageChan <- m
			}

		} else {
			log.Printf("INFO: no messages available")
		}
	}
}

//
// end of file
//
