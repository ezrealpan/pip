package models

import (
	"sync"

	"ezreal.com.cn/pip/pip"
)

// RunningOutput contains the output configuration
type RunningOutput struct {
	Output pip.Output
	// Must be 64-bit aligned
	newMetricsCount int64
	droppedMetrics  int64

	aggMutex sync.Mutex
}

// Init ...
func (r *RunningOutput) Init() error {
	if p, ok := r.Output.(pip.Initializer); ok {
		return p.Init()
	}
	return nil
}

// NewRunningOutput ....
func NewRunningOutput(output pip.Output) *RunningOutput {
	return &RunningOutput{
		Output: output,
	}
}

// AddMetric ...
// AddMetric adds a metric to the output.
//
// Takes ownership of metric
func (r *RunningOutput) AddMetric(metric pip.Metric) {
	r.Output.Write([]pip.Metric{metric})
}
