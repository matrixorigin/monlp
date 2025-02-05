// agent package defines the interface for the agent, which is the basic unit of the pipeline
// or workflow.   Each agent process input data and produce output data, then user can connect
// multiple agents to form a pipeline.
// The both input and output are json, with an optional desc field and a data field.
//
//	{
//	    "desc": An optional description json object
//	    "data": A json object
//	}
//
// The data and desc json object are agent specific and should have been documented by each
// agent.
package agent

import (
	"fmt"
	"iter"
)

var (
	ErrExecOneNA   = fmt.Errorf("ExecuteOne not implemented")
	ErrYieldDone   = fmt.Errorf("Yield done")
	ErrStreamBreak = fmt.Errorf("Stream break")
)

type Agent interface {
	// Config the agent with a config file
	Config(bs []byte) error
	// Execute the agent with input data, return the output data and error
	Execute(it iter.Seq2[[]byte, error], dict map[string]string) (iter.Seq2[[]byte, error], error)
	// ExecuteOne
	ExecuteOne(input []byte, dict map[string]string, yield func([]byte, error) bool) error
	// Close the agent
	Close() error
	// Set a value
	SetValue(key string, value interface{}) error
}

type NilKVAgent struct {
}

func (na *NilKVAgent) SetValue(key string, value interface{}) error {
	return fmt.Errorf("Agent setting unknown value %s=%v", key, value)
}

type NilConfigAgent struct {
}

func (na *NilConfigAgent) Config(bs []byte) error {
	return nil
}

type NilCloseAgent struct {
}

func (na *NilCloseAgent) Close() error {
	return nil
}

type AgentPipe struct {
	NilKVAgent
	NilCloseAgent
	agents []Agent
}

func (ap *AgentPipe) AddAgent(agent Agent) {
	ap.agents = append(ap.agents, agent)
}

func (ap *AgentPipe) Execute(it iter.Seq2[[]byte, error], dict map[string]string) (iter.Seq2[[]byte, error], error) {
	var err error
	for _, agent := range ap.agents {
		it, err = agent.Execute(it, dict)
		if err != nil {
			return nil, err
		}
	}
	return it, nil
}
func (ap *AgentPipe) ExecuteOne(input []byte, dict map[string]string, yield func([]byte, error) bool) error {
	return ErrExecOneNA
}

func (ap *AgentPipe) Close() error {
	// close ALL, even if we have error in the middle.
	// Only return the first error
	var savedErr error
	for _, agent := range ap.agents {
		err := agent.Close()
		if err != nil {
			savedErr = err
		}
	}
	return savedErr
}

type AgentFanOut struct {
	NilKVAgent
	NilConfigAgent
	Agents []Agent
}

func (af *AgentFanOut) Execute(input iter.Seq2[[]byte, error], dict map[string]string) (iter.Seq2[[]byte, error], error) {
	var err error
	for _, agent := range af.Agents {
		_, err = agent.Execute(input, dict)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (af *AgentFanOut) ExecuteOne(input []byte, dict map[string]string, yield func([]byte, error) bool) error {
	return ErrExecOneNA
}

func (af *AgentFanOut) Close() error {
	// close ALL, even if we have error in the middle.
	// Only return the first error
	var savedErr error
	for _, agent := range af.Agents {
		err := agent.Close()
		if err != nil {
			savedErr = err
		}
	}
	return savedErr
}

type SimpleExecuteAgent struct {
	Self Agent
}

func (sa *SimpleExecuteAgent) Execute(input iter.Seq2[[]byte, error], dict map[string]string) (iter.Seq2[[]byte, error], error) {
	return func(yield func([]byte, error) bool) {
		if input == nil {
			return
		}

		for data, err := range input {
			if err != nil {
				yield(nil, err)
				return
			}

			err = sa.Self.ExecuteOne(data, dict, yield)
			if err == ErrYieldDone {
				return
			} else if err != nil {
				if !yield(nil, err) {
					return
				}
			}
		}
	}, nil
}

type StringArrayAgent struct {
	NilConfigAgent
	NilCloseAgent
	values []string
}

func NewStringArrayAgent(values []string) *StringArrayAgent {
	sa := &StringArrayAgent{}
	sa.values = values
	return sa
}

func (sa *StringArrayAgent) SetValue(name string, value interface{}) error {
	var ok bool
	sa.values, ok = value.([]string)
	if !ok {
		return fmt.Errorf("StringArrayAgent: SetValue: value is not []string")
	}
	return nil
}

func (sa *StringArrayAgent) ExecuteOne(input []byte, dict map[string]string, yield func([]byte, error) bool) error {
	return ErrExecOneNA
}

func (sa *StringArrayAgent) Execute(input iter.Seq2[[]byte, error], dict map[string]string) (iter.Seq2[[]byte, error], error) {
	return func(yield func([]byte, error) bool) {
		for _, data := range sa.values {
			if !yield([]byte(data), nil) {
				return
			}
		}
	}, nil
}
