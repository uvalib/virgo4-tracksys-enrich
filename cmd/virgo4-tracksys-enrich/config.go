package main

import (
	"log"
	"os"
	"strconv"
)

// ServiceConfig defines all of the service configuration parameters
type ServiceConfig struct {
	Mode string // are we in "sirsi" mode or "other" mode

	InQueueName       string // SQS queue name for inbound documents
	OutQueueName      string // SQS queue name for outbound documents
	PollTimeOut       int64  // the SQS queue timeout (in seconds)
	MessageBucketName string // the bucket to use for large messages

	ServiceEndpoint string // the URL of the tracksys endpoint
	ServiceTimeout  int    // service timeout (in seconds)

	CacheLoadApi    string // the API path used to populate the cache
	CacheDetailsApi string // the API path used to get the cache details
	CacheAge        int    // how frequently do we reload the cache (in seconds)

	PidDetailsApi  string // the API path used to get the OCR eligible info
	OcrServiceRoot string // the root link for the OCR service (for eligible items)

	RewriteFields map[string]string // the fields we explicitly rewrite

	RightsEndpoint string // the endpoint for getting use policy (as part of the enrichment process)
	OembedRoot     string // the oembed url root

	DigitalContentCacheRoot   string // the root url of the digital content cache
	DigitalContentCacheBucket string // the name of the bucket for the digital content cache

	WorkerQueueSize int // the inbound message queue size to feed the workers
	Workers         int // the number of worker processes
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

func envToInt(env string) int {

	number := ensureSetAndNonEmpty(env)
	n, err := strconv.Atoi(number)
	if err != nil {

		os.Exit(1)
	}
	return n
}

// LoadConfiguration will load the service configuration from env/cmdline
// and return a pointer to it. Any failures are fatal.
func LoadConfiguration() *ServiceConfig {

	var cfg ServiceConfig

	//
	// Sirsi mode means we are processing Sirsi items in the pipeline. These are
	// identified by the unique Sirsi catalog key (catkey) and have zero or more
	// digital items available in tracksys
	//
	// Non-sirsi mode means we are processing other kinds of digital items. These
	// are identified by the unique "part ID" (pid) and have zero or one digital
	// items available in tracksys.

	// we use different tracksys API endpoints to process the different types

	cfg.Mode = ensureSetAndNonEmpty("VIRGO4_TRACKSYS_ENRICH_MODE")

	cfg.InQueueName = ensureSetAndNonEmpty("VIRGO4_TRACKSYS_ENRICH_IN_QUEUE")
	cfg.OutQueueName = ensureSetAndNonEmpty("VIRGO4_TRACKSYS_ENRICH_OUT_QUEUE")
	cfg.MessageBucketName = ensureSetAndNonEmpty("VIRGO4_SQS_MESSAGE_BUCKET")
	cfg.PollTimeOut = int64(envToInt("VIRGO4_TRACKSYS_ENRICH_QUEUE_POLL_TIMEOUT"))

	cfg.ServiceEndpoint = ensureSetAndNonEmpty("VIRGO4_TRACKSYS_ENRICH_SERVICE_URL")
	cfg.ServiceTimeout = envToInt("VIRGO4_TRACKSYS_ENRICH_SERVICE_TIMEOUT")

	cfg.RightsEndpoint = ensureSetAndNonEmpty("VIRGO4_TRACKSYS_ENRICH_RIGHTS_URL")
	cfg.OembedRoot = ensureSetAndNonEmpty("VIRGO4_TRACKSYS_ENRICH_OEMBED_ROOT")

	cfg.WorkerQueueSize = envToInt("VIRGO4_TRACKSYS_ENRICH_WORK_QUEUE_SIZE")
	cfg.Workers = envToInt("VIRGO4_TRACKSYS_ENRICH_WORKERS")

	cfg.DigitalContentCacheRoot = ensureSetAndNonEmpty("VIRGO4_TRACKSYS_ENRICH_CACHE_ROOT_URL")
	cfg.DigitalContentCacheBucket = ensureSetAndNonEmpty("VIRGO4_TRACKSYS_ENRICH_CACHE_BUCKET")

	cfg.CacheLoadApi = ensureSetAndNonEmpty("VIRGO4_TRACKSYS_ENRICH_CACHE_LOAD")
	cfg.CacheDetailsApi = ensureSetAndNonEmpty("VIRGO4_TRACKSYS_ENRICH_CACHE_DETAILS")
	cfg.CacheAge = envToInt("VIRGO4_TRACKSYS_ENRICH_CACHE_AGE")

	cfg.PidDetailsApi = ensureSetAndNonEmpty("VIRGO4_TRACKSYS_ENRICH_PID_DETAILS")
	cfg.OcrServiceRoot = ensureSetAndNonEmpty("VIRGO4_TRACKSYS_ENRICH_OCR_SERVICE_ROOT")

	// maybe configure later
	cfg.RewriteFields = map[string]string{"uva_availability_f_stored": "Online", "anon_availability_f_stored": "Online"}

	log.Printf("[CONFIG] Mode                      = [%s]", cfg.Mode)
	log.Printf("[CONFIG] InQueueName               = [%s]", cfg.InQueueName)
	log.Printf("[CONFIG] OutQueueName              = [%s]", cfg.OutQueueName)
	log.Printf("[CONFIG] PollTimeOut               = [%d]", cfg.PollTimeOut)
	log.Printf("[CONFIG] MessageBucketName         = [%s]", cfg.MessageBucketName)

	log.Printf("[CONFIG] ServiceEndpoint           = [%s]", cfg.ServiceEndpoint)
	log.Printf("[CONFIG] ServiceTimeout            = [%d]", cfg.ServiceTimeout)
	log.Printf("[CONFIG] CacheLoadApi              = [%s]", cfg.CacheLoadApi)
	log.Printf("[CONFIG] CacheDetailsApi           = [%s]", cfg.CacheDetailsApi)
	log.Printf("[CONFIG] CacheAge                  = [%d]", cfg.CacheAge)

	log.Printf("[CONFIG] PidDetailsApi             = [%s]", cfg.PidDetailsApi)
	log.Printf("[CONFIG] OcrServiceRoot            = [%s]", cfg.OcrServiceRoot)

	log.Printf("[CONFIG] RightsEndpoint            = [%s]", cfg.RightsEndpoint)
	log.Printf("[CONFIG] OembedRoot                = [%s]", cfg.OembedRoot)

	log.Printf("[CONFIG] DigitalContentCacheRoot   = [%s]", cfg.DigitalContentCacheRoot)
	log.Printf("[CONFIG] DigitalContentCacheBucket = [%s]", cfg.DigitalContentCacheBucket)

	log.Printf("[CONFIG] WorkerQueueSize           = [%d]", cfg.WorkerQueueSize)
	log.Printf("[CONFIG] Workers                   = [%d]", cfg.Workers)

	return &cfg
}

//
// end of file
//
