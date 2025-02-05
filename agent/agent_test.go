package agent

import (
	"strings"
	"testing"

	"github.com/matrixorigin/monlp/common"
)

type dupAgent struct {
	NilKVAgent
	NilConfigAgent
	NilCloseAgent
	SimpleExecuteAgent
}

func newDupAgent() *dupAgent {
	da := &dupAgent{}
	da.Self = da
	return da
}

func (da *dupAgent) ExecuteOne(data []byte, dict map[string]string, yield func([]byte, error) bool) error {
	if !yield(data, nil) {
		return ErrYieldDone
	}

	data2 := make([]byte, len(data)*2)
	copy(data2, data)
	copy(data2[len(data):], data)
	if !yield(data2, nil) {
		return ErrYieldDone
	}

	return nil
}

type dropAgent struct {
	NilKVAgent
	NilConfigAgent
	NilCloseAgent
	SimpleExecuteAgent
	prefix string
}

func newDropAgent(prefix string) *dropAgent {
	da := &dropAgent{prefix: prefix}
	da.Self = da
	return da
}

func (da *dropAgent) ExecuteOne(data []byte, dict map[string]string, yield func([]byte, error) bool) error {
	if !strings.HasPrefix(string(data), da.prefix) {
		if !yield(data, nil) {
			return ErrYieldDone
		}
	}
	return nil
}

func TestAgent(t *testing.T) {
	sa := NewStringArrayAgent([]string{
		"cat",
		"dog",
		"foo bar",
		"duck",
	})

	dupa := newDupAgent()
	dropa := newDropAgent("foo")

	var pipe AgentPipe
	pipe.AddAgent(sa)
	pipe.AddAgent(dupa)
	pipe.AddAgent(dropa)

	defer pipe.Close()

	pipeline, err := pipe.Execute(nil, nil)
	common.Assert(t, err == nil, "pipe.Execute failed")

	var res []string
	for s, err := range pipeline {
		common.Assert(t, err == nil, "pipeline failed")
		res = append(res, string(s))
	}

	common.Assert(t, len(res) == 6, "Expected 6, got %v", len(res))
	common.Assert(t, res[0] == "cat", "Expected cat, got %v", res[0])
	common.Assert(t, res[1] == "catcat", "Expected cat, got %v", res[1])
	common.Assert(t, res[2] == "dog", "Expected dog, got %v", res[2])
	common.Assert(t, res[3] == "dogdog", "Expected dog, got %v", res[3])
	common.Assert(t, res[4] == "duck", "Expected duck, got %v", res[4])
	common.Assert(t, res[5] == "duckduck", "Expected duck, got %v", res[5])
}
