package main

import "github.com/uvalib/virgo4-sqs-sdk/awssqs"

// PipelineStep - the interface representing a processing step of the enrich pipeline
type PipelineStep interface {

	// the name of the step
	Name() string

	// process the provided message and return:
	//    bool  - should the pipeline continue to the next step?
	//    bool  - was the message updated during this step?
	//    error - did an error occur?
	Process(*awssqs.Message) (bool, bool, error)
}

//// Pipeline - the interface representing the complete enrich pipeline
//type Pipeline interface {
//
//	// process the provided message and return:
//	//    int   - the step that failed or -1 if successful
//	//    error - did an error occur?
//	Process(*awssqs.Message) (int, error)
//}
//
//// this is our actual pipeline implementation
//type pipelineImpl struct {
//	steps []PipelineStep // the individual steps of the enrich pipeline
//}
//
//// NewPipeline - the factory for the enrich pipeline
//func NewPipeline(config *ServiceConfig) Pipeline {
//
//	// mock implementation here if necessary
//
//	impl := &pipelineImpl{}
//	impl.steps = make([]PipelineStep, 0)
//
//	// the pipeline consists of 3 steps:
//	//  1. tracksys enrich
//	//  2. field rewrite
//	//  3. xxx
//
//	return impl
//}
//
//func (p *pipelineImpl) Process(message *awssqs.Message) (int, error) {
//	return -1, nil
//}

//
// end of file
//
