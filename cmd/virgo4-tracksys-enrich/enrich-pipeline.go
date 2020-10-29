package main

import (
	"github.com/uvalib/virgo4-sqs-sdk/awssqs"
	"log"
)

// PipelineStep - the interface representing a processing step of the enrich pipeline
type PipelineStep interface {

	// the name of the step
	Name() string

	// process the provided message and return:
	//    bool        - should the pipeline continue to the next step?
	//    interface{} - anything to be passed to the next step
	//    error       - did an error occur?
	Process(*awssqs.Message, interface{}) (bool, interface{}, error)
}

// Pipeline - the interface representing the complete enrich pipeline
type Pipeline interface {

	// process the provided message and return:
	//    int   - the step that failed or -1 if successful
	//    error - did an error occur?
	Process(*awssqs.Message) (int, error)
}

// this is our actual pipeline implementation
type pipelineImpl struct {
	steps []PipelineStep // the individual steps of the enrich pipeline
}

// NewEnrichPipeline - the factory for the enrich pipeline
func NewEnrichPipeline(config *ServiceConfig) Pipeline {

	// mock implementation here if necessary

	impl := &pipelineImpl{}
	impl.steps = make([]PipelineStep, 0)

	// the pipeline consists of 4 steps:
	//  1. tracksys extract step
	//  2. tracksys enrich step
	//  3. field rewrite step
	//  4. partial digitization step
	//  5. metadata cache step

	impl.steps = append(impl.steps, NewTracksysExtractStep(config))
	impl.steps = append(impl.steps, NewTracksysEnrichStep(config))
	impl.steps = append(impl.steps, NewFieldRewriteStep(config))
	impl.steps = append(impl.steps, NewPartialDigitizedStep(config))
	impl.steps = append(impl.steps, NewMetaDataCacheStep(config))

	return impl
}

func (pi *pipelineImpl) Process(message *awssqs.Message) (int, error) {

	var payload interface{}
	for ix, step := range pi.steps {
		doNext, data, err := step.Process(message, payload)

		// error happened during a step
		if err != nil {
			log.Printf("WARNING: enrich pipeline failed at step %d (%s)", ix, step.Name())
			// return step number and error
			return ix, err
		}

		// no error but don't continue the pipeline
		if doNext == false {
			//log.Printf("INFO: enrich pipeline exited early at step %d (%s)", ix, step.Name())
			// all is well
			return -1, nil
		}

		// to pass on to the next step
		payload = data
	}

	// done all the steps and all is well
	return -1, nil
}

//
// end of file
//
