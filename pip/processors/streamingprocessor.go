package processors

import (
	"ezreal.com.cn/pip/pip"
)

// NewStreamingProcessorFromProcessor is a converter that turns a standard
// processor into a streaming processor
func NewStreamingProcessorFromProcessor(p pip.Processor) pip.StreamingProcessor {
	sp := &streamingProcessor{
		processor: p,
	}
	return sp
}

type streamingProcessor struct {
	processor pip.Processor
	acc       pip.Accumulator
}

func (sp *streamingProcessor) SampleConfig() string {
	return sp.processor.SampleConfig()
}

func (sp *streamingProcessor) Description() string {
	return sp.processor.Description()
}

func (sp *streamingProcessor) Start(acc pip.Accumulator) error {

	sp.acc = acc
	return nil
}

func (sp *streamingProcessor) Add(m pip.Metric, acc pip.Accumulator) error {
	for _, m := range sp.processor.Apply(m) {
		acc.AddMetric(m)
	}
	return nil
}

func (sp *streamingProcessor) Stop() error {
	return nil
}

// Make the streamingProcessor of type Initializer to be able
// to call the Init method of the wrapped processor if
// needed
func (sp *streamingProcessor) Init() error {
	if p, ok := sp.processor.(pip.Initializer); ok {
		err := p.Init()
		if err != nil {
			return err
		}
	}
	return nil
}

// Unwrap lets you retrieve the original telegraf.Processor from the
// StreamingProcessor. This is necessary because the toml Unmarshaller won't
// look inside composed types.
func (sp *streamingProcessor) Unwrap() pip.Processor {
	return sp.processor
}
