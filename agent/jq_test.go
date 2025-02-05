package agent

import (
	"testing"

	"github.com/matrixorigin/monlp/common"
)

func TestJqAgent(t *testing.T) {
	stra := NewStringArrayAgent([]string{
		`{"foo": 111, "bar": 222}`,
		`{"foo": "bar", "bar": "zoo"}`,
		`{"data": {"a": 1, "b": 2}}`,
		`{"data": {"a": 3, "b": "bbb"}}`,
	})

	jqa, err := NewJqAgent("")
	common.Assert(t, err == nil, "Expected nil, got %v", err)

	var pipe AgentPipe
	pipe.AddAgent(stra)
	pipe.AddAgent(jqa)
	defer pipe.Close()

	err = jqa.SetValue("jq", ".foo")
	common.Assert(t, err == nil, "Expected nil, got %v", err)

	it, err := pipe.Execute(nil, nil)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	for data, err := range it {
		common.Assert(t, err == nil, "Expected nil, got %v", err)
		t.Logf("Query Result: %s", string(data))
	}
}
