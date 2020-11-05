package agent

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"ezreal.com.cn/pip/config"
	"ezreal.com.cn/pip/pip"
	"ezreal.com.cn/pip/pip/models"
)

// Agent runs a set of plugins.
type Agent struct {
	Config *config.Config
}

// NewAgent returns an Agent for the given Config.
func NewAgent(config *config.Config) (*Agent, error) {
	a := &Agent{
		Config: config,
	}
	return a, nil
}

// inputUnit is a group of input plugins and the shared channel they write to.
//
// ┌───────┐
// │ Input │───┐
// └───────┘   │
// ┌───────┐   │     ______
// │ Input │───┼──▶ ()_____)
// └───────┘   │
// ┌───────┐   │
// │ Input │───┘
// └───────┘
type inputUnit struct {
	dst    chan<- pip.Metric
	inputs []*models.RunningInput
}

//  ______     ┌───────────┐     ______
// ()_____)──▶ │ Processor │──▶ ()_____)
//             └───────────┘
type processorUnit struct {
	src       <-chan pip.Metric
	dst       chan<- pip.Metric
	processor *models.RunningProcessor
}

// outputUnit is a group of Outputs and their source channel.  pip.Metrics on the
// channel are written to all outputs.
//
//                            ┌────────┐
//                       ┌──▶ │ Output │
//                       │    └────────┘
//  ______     ┌─────┐   │    ┌────────┐
// ()_____)──▶ │ Fan │───┼──▶ │ Output │
//             └─────┘   │    └────────┘
//                       │    ┌────────┐
//                       └──▶ │ Output │
//                            └────────┘
type outputUnit struct {
	src     <-chan pip.Metric
	outputs []*models.RunningOutput
}

// Run starts and runs the Agent until the context is done.
func (a *Agent) Run(ctx context.Context) error {
	err := a.initPlugins()
	if err != nil {
		return err
	}

	startTime := time.Now()
	next, ou, err := a.startOutputs(ctx, a.Config.Outputs)
	if err != nil {
		return err
	}

	var pu []*processorUnit
	if len(a.Config.Processors) != 0 {
		next, pu, err = a.startProcessors(next, a.Config.Processors)
		if err != nil {
			return err
		}
	}

	iu, err := a.startInputs(next, a.Config.Inputs)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = a.runOutputs(ou)
		if err != nil {
			log.Printf("E! [agent] Error running outputs: %v", err)
		}
	}()

	if pu != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err = a.runProcessors(pu)
			if err != nil {
				log.Printf("E! [agent] Error running processors: %v", err)
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = a.runInputs(ctx, startTime, iu)
		if err != nil {
			log.Printf("E! [agent] Error running inputs: %v", err)
		}
	}()

	wg.Wait()
	log.Printf("D! [agent] Stopped Successfully")
	return err
}

// initPlugins runs the Init function on plugins.
func (a *Agent) initPlugins() error {
	for _, input := range a.Config.Inputs {
		err := input.Init()
		if err != nil {
			return err
		}
	}
	for _, processor := range a.Config.Processors {
		err := processor.Init()
		if err != nil {
			return err
		}
	}

	for _, output := range a.Config.Outputs {
		err := output.Init()
		if err != nil {
			return err
		}
	}
	return nil
}

// startOutputs calls Connect on all outputs and returns the source channel.
// If an error occurs calling Connect all stared plugins have Close called.
func (a *Agent) startOutputs(
	ctx context.Context,
	outputs []*models.RunningOutput,
) (chan<- pip.Metric, *outputUnit, error) {
	src := make(chan pip.Metric, 100)

	unit := &outputUnit{src: src, outputs: outputs}
	return src, unit, nil
}

// startProcessors sets up the processor chain and calls Start on all
// processors.  If an error occurs any started processors are Stopped.
func (a *Agent) startProcessors(
	dst chan<- pip.Metric,
	processors models.RunningProcessors,
) (chan<- pip.Metric, []*processorUnit, error) {
	var units []*processorUnit

	var src chan pip.Metric
	for _, processor := range processors {
		src = make(chan pip.Metric, 100)
		acc := NewAccumulator(processor, dst)

		err := processor.Start(acc)
		if err != nil {
			for _, u := range units {
				u.processor.Stop()
				close(u.dst)
			}
			return nil, nil, fmt.Errorf("starting processor %s: %w", processor.LogName(), err)
		}

		units = append(units, &processorUnit{
			src:       src,
			dst:       dst,
			processor: processor,
		})

		dst = src
	}

	return src, units, nil
}

func (a *Agent) startInputs(
	dst chan<- pip.Metric,
	inputs []*models.RunningInput,
) (*inputUnit, error) {
	unit := &inputUnit{
		dst:    dst,
		inputs: inputs,
	}
	return unit, nil
}

// runOutputs begins processing pip.metrics and returns until the source channel is
// closed and all pip.metrics have been written.  On shutdown pip.metrics will be
// written one last time and dropped if unsuccessful.
func (a *Agent) runOutputs(
	unit *outputUnit,
) error {
	for metric := range unit.src {
		for _, output := range unit.outputs {
			output.AddMetric(metric)
		}
	}

	log.Println("I! [agent] Hang on, flushing any cached metrics before shutdown")
	return nil
}

// runProcessors begins processing pip.metrics and runs until the source channel is
// closed and all pip.metrics have been written.
func (a *Agent) runProcessors(
	units []*processorUnit,
) error {
	var wg sync.WaitGroup
	for _, unit := range units {
		wg.Add(1)
		go func(unit *processorUnit) {
			defer wg.Done()

			acc := NewAccumulator(unit.processor, unit.dst)
			for m := range unit.src {

				err := unit.processor.Add(m, acc)
				fmt.Printf("runProcessors %+v, err %+v\n", m, err)
				if err != nil {
					acc.AddError(err)
					m.Drop()
				}
			}
			unit.processor.Stop()
			close(unit.dst)
			log.Printf("D! [agent] Processor channel closed")
		}(unit)
	}
	wg.Wait()

	return nil
}

// runInputs starts and triggers the periodic gather for Inputs.
//
// When the context is done the timers are stopped and this function returns
// after all ongoing Gather calls complete.
func (a *Agent) runInputs(
	ctx context.Context,
	startTime time.Time,
	unit *inputUnit,
) error {
	for _, input := range unit.inputs {
		acc := NewAccumulator(input, unit.dst)

		for {
			time.Sleep(3 * time.Second)
			input.Input.Gather(acc)
		}
	}
	return nil
}

// gather runs an input's gather function periodically until the context is
// done.
func (a *Agent) gatherLoop(
	ctx context.Context,
	acc pip.Accumulator,
	input *models.RunningInput,
) {

}
