package main

import (
	"log"
	"os"
	"strconv"
)

// ServiceConfig defines all of the service configuration parameters
type ServiceConfig struct {
	InQueueName    string    // SQS queue name for inbound documents
	OutQueue1Name  string    // SQS queue name for outbound documents
	OutQueue2Name  string    // SQS queue name for outbound documents
	PollTimeOut    int64     // the SQS queue timeout (in seconds)

	WorkerQueueSize int      // the inbound message queue size to feed the workers
	Workers         int      // the number of worker processes
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

	cfg.InQueueName = ensureSetAndNonEmpty( "VIRGO4_SQS_FORK_IN_QUEUE" )
	cfg.OutQueue1Name = ensureSetAndNonEmpty( "VIRGO4_SQS_FORK_OUT_1_QUEUE" )
	cfg.OutQueue2Name = ensureSetAndNonEmpty( "VIRGO4_SQS_FORK_OUT_2_QUEUE" )
	cfg.PollTimeOut = int64( envToInt( "VIRGO4_SQS_FORK_QUEUE_POLL_TIMEOUT" ) )
	cfg.WorkerQueueSize = envToInt( "VIRGO4_SQS_FORK_WORK_QUEUE_SIZE" )
	cfg.Workers = envToInt( "VIRGO4_SQS_FORK_WORKERS" )

	log.Printf("[CONFIG] InQueueName          = [%s]", cfg.InQueueName )
	log.Printf("[CONFIG] OutQueue1Name        = [%s]", cfg.OutQueue1Name )
	log.Printf("[CONFIG] OutQueue2Name        = [%s]", cfg.OutQueue2Name )
	log.Printf("[CONFIG] PollTimeOut          = [%d]", cfg.PollTimeOut )
	log.Printf("[CONFIG] WorkerQueueSize      = [%d]", cfg.WorkerQueueSize )
	log.Printf("[CONFIG] Workers              = [%d]", cfg.Workers )

	return &cfg
}

//
// end of file
//