package models

import "ezreal.com.cn/pip/pip"

// RunningInput ...
type RunningInput struct {
	Input pip.Input
}

// LogName ...
func (r *RunningInput) LogName() string {
	return ""
}

// MakeMetric ...
func (r *RunningInput) MakeMetric(metric pip.Metric) pip.Metric {
	return metric
}

// Log ...
func (r *RunningInput) Log() pip.Logger {
	return nil
}

// Init ...
func (r *RunningInput) Init() error {
	if p, ok := r.Input.(pip.Initializer); ok {
		return p.Init()
	}
	return nil
}

// NewRunningInput ...
func NewRunningInput(input pip.Input) *RunningInput {
	return &RunningInput{
		Input: input,
	}
}
