package main

import (
	"log"
	"os"
	"strconv"
)

// ServiceConfig defines all of the service configuration parameters
type ServiceConfig struct {
	InQueueName       string   // SQS queue name for inbound documents
	OutQueueName      string   // SQS queue name for outbound documents
	PollTimeOut       int64    // the SQS queue timeout (in seconds)
	MessageBucketName string   // the bucket to use for large messages

	ServiceEndpoint   string   // the URL of the tracksys endpoint
	ServiceTimeout    int64    // service timeout (in seconds)

	WorkerQueueSize   int      // the inbound message queue size to feed the workers
	Workers           int      // the number of worker processes
}

func ensureSet(env string) string {
	val, set := os.LookupEnv(env)

	if set == false {
		log.Printf("environment variable not set: [%s]", env)
		os.Exit(1)
	}

	return val
}

func ensureSetAndNonEmpty(env string) string {
	val := ensureSet(env)

	if val == "" {
		log.Printf("environment variable not set: [%s]", env)
		os.Exit(1)
	}

	return val
}

func envToInt( env string ) int {

	number := ensureSetAndNonEmpty( env )
	n, err := strconv.Atoi( number )
	if err != nil {

		os.Exit(1)
	}
	return n
}

// LoadConfiguration will load the service configuration from env/cmdline
// and return a pointer to it. Any failures are fatal.
func LoadConfiguration() *ServiceConfig {

	var cfg ServiceConfig

	cfg.InQueueName = ensureSetAndNonEmpty( "VIRGO4_TRACKSYS_ENRICH_IN_QUEUE" )
	cfg.OutQueueName = ensureSetAndNonEmpty( "VIRGO4_TRACKSYS_ENRICH_OUT_QUEUE" )
	cfg.MessageBucketName = ensureSetAndNonEmpty( "VIRGO4_SQS_MESSAGE_BUCKET" )
	cfg.PollTimeOut = int64( envToInt( "VIRGO4_TRACKSYS_ENRICH_QUEUE_POLL_TIMEOUT" ) )

	cfg.ServiceEndpoint = ensureSetAndNonEmpty( "VIRGO4_TRACKSYS_ENRICH_SERVICE_URL" )
	cfg.ServiceTimeout = int64( envToInt( "VIRGO4_TRACKSYS_ENRICH_SERVICE_TIMEOUT" ) )

	cfg.WorkerQueueSize = envToInt( "VIRGO4_TRACKSYS_ENRICH_WORK_QUEUE_SIZE" )
	cfg.Workers = envToInt( "VIRGO4_TRACKSYS_ENRICH_WORKERS" )

	log.Printf("[CONFIG] InQueueName          = [%s]", cfg.InQueueName )
	log.Printf("[CONFIG] OutQueueName         = [%s]", cfg.OutQueueName )
	log.Printf("[CONFIG] PollTimeOut          = [%d]", cfg.PollTimeOut )
	log.Printf("[CONFIG] MessageBucketName    = [%s]", cfg.MessageBucketName )

	log.Printf("[CONFIG] ServiceEndpoint      = [%s]", cfg.ServiceEndpoint )
	log.Printf("[CONFIG] ServiceTimeout       = [%d]", cfg.ServiceTimeout )

	log.Printf("[CONFIG] WorkerQueueSize      = [%d]", cfg.WorkerQueueSize )
	log.Printf("[CONFIG] Workers              = [%d]", cfg.Workers )

	return &cfg
}

//
// end of file
//