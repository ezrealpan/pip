package models

import "ezreal.com.cn/pip/pip"

// RunningInput ...
type RunningInput struct {
	Input pip.Input

	Config *InputConfig

	defaultTags map[string]string
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
func NewRunningInput(input pip.Input, config *InputConfig) *RunningInput {
	return &RunningInput{
		Input:  input,
		Config: config,
	}
}

// InputConfig is the common config for all inputs.
type InputConfig struct {
	Name string
	Tags map[string]string
}

// SetDefaultTags ...
func (r *RunningInput) SetDefaultTags(tags map[string]string) {
	r.defaultTags = tags
}
