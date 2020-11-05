package printer

import (
	"fmt"

	"ezreal.com.cn/pip/pip"
	"ezreal.com.cn/pip/pip/processors"
)

// Printer ...
type Printer struct {
}

var sampleConfig = `
`

// SampleConfig ...
func (p *Printer) SampleConfig() string {
	return sampleConfig
}

// Description ...
func (p *Printer) Description() string {
	return "Print all metrics that pass through this filter."
}

// Init ...
func (p *Printer) Init() error {
	fmt.Printf("Simple Printer Init %+v\n", "----------------------------------->")
	return nil
}

// Apply ...
func (p *Printer) Apply(in ...pip.Metric) []pip.Metric {
	for _, metric := range in {
		fmt.Printf("Processor %+v", metric)
	}
	return in
}

func init() {
	processors.Add("printer", func() pip.Processor {
		return &Printer{}
	})
}
