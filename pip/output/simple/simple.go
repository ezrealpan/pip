package simple

import (
	"fmt"

	"ezreal.com.cn/pip/pip"
	"ezreal.com.cn/pip/pip/output"
)

// Simple ...
type Simple struct {
	Ok bool `toml:"ok"`
}

// Description ...
func (s *Simple) Description() string {
	return "a demo output"
}

// SampleConfig ...
func (s *Simple) SampleConfig() string {
	return `
  ok = true
`
}

// Init ...
func (s *Simple) Init() error {
	fmt.Printf("Simple Output Init %+v\n", "----------------------------------->")
	return nil
}

// Connect ...
func (s *Simple) Connect() error {
	// Make a connection to the URL here
	return nil
}

// Close ...
func (s *Simple) Close() error {
	// Close connection to the URL here
	return nil
}

// Write ...
func (s *Simple) Write(metrics []pip.Metric) error {
	for _, metric := range metrics {
		fmt.Println("output Write field", metric.Fields())
		fmt.Printf("output Write %+v\n", metric)
	}
	return nil
}

func init() {
	output.Add("simpleoutput", func() pip.Output { return &Simple{} })
}
