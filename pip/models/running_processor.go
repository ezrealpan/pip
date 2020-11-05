package models

import (
	"fmt"
	"sync"

	"ezreal.com.cn/pip/pip"
)

// ProcessorConfig ...
// FilterConfig containing a name and filter
type ProcessorConfig struct {
	Name  string
	Alias string
	Order int64
}

// RunningProcessor ...
type RunningProcessor struct {
	Processor pip.StreamingProcessor
	sync.Mutex
}

// RunningProcessors ...
type RunningProcessors []*RunningProcessor

// NewRunningProcessor ...
func NewRunningProcessor(processor pip.StreamingProcessor) *RunningProcessor {

	return &RunningProcessor{
		Processor: processor,
	}
}

// Init ...
func (r *RunningProcessor) Init() error {
	if p, ok := r.Processor.(pip.Initializer); ok {
		return p.Init()
	}
	return nil
}

// Log ...
func (r *RunningProcessor) Log() pip.Logger {
	return nil
}

// LogName ...
func (r *RunningProcessor) LogName() string {
	return ""
}

// MakeMetric ...
func (r *RunningProcessor) MakeMetric(metric pip.Metric) pip.Metric {
	return metric
}

func (r *RunningProcessor) Start(acc pip.Accumulator) error {
	fmt.Printf("RunningProcessor %+v ----------------->\n", "Start")
	return r.Processor.Start(acc)
}

func (r *RunningProcessor) Stop() {
	r.Processor.Stop()
}

func (r *RunningProcessor) Add(m pip.Metric, acc pip.Accumulator) error {
	// if ok := r.Config.Filter.Select(m); !ok {
	// 	// pass downstream
	// 	acc.AddMetric(m)
	// 	return nil
	// }

	// r.Config.Filter.Modify(m)
	// if len(m.FieldList()) == 0 {
	// 	// drop metric
	// 	r.metricFiltered(m)
	// 	return nil
	// }

	return r.Processor.Add(m, acc)
}
