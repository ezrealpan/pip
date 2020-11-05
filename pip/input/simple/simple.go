package simple

import (
	"fmt"
	"time"

	"ezreal.com.cn/pip/pip"
	"ezreal.com.cn/pip/pip/input"
)

// Simple ...
type Simple struct {
	Ok  bool   `toml:"ok"`
	Tip string `toml:"tip"`
}

// Description ...
func (s *Simple) Description() string {
	return "a demo plugin"
}

// SampleConfig ...
func (s *Simple) SampleConfig() string {
	return `
  ## Indicate if everything is fine
  ok = true
`
}

// Init ...
func (s *Simple) Init() error {
	fmt.Printf("Simple Init %+v\n", "----------------------------------->")
	fmt.Println("simple.tip", s.Tip)
	return nil
}

// Gather ...
func (s *Simple) Gather(acc pip.Accumulator) error {

	t := time.Now().Format("Mon Jan 2 15:04:05 -0700 MST 2006")
	acc.AddFields(t, map[string]interface{}{"value": "pretty good"}, nil)

	return nil
}

func init() {
	input.Add("simple", func() pip.Input { return &Simple{} })
}
