package wikix

import (
	"bufio"
	"encoding/json"
	"testing"

	"github.com/matrixorigin/monlp/agent"
	"github.com/matrixorigin/monlp/agent/llm"
	"github.com/matrixorigin/monlp/common"
)

func TestWikiSimple(t *testing.T) {
	qfile, err := common.OpenFileForTest("agent", "wikix", "questions.jsonl")
	common.Assert(t, err == nil, "OpenFileForTest failed: %v", err)
	// next read qfile line by line
	var qlines, lines []string
	scanner := bufio.NewScanner(qfile)
	for scanner.Scan() {
		qline := scanner.Text()
		qlines = append(qlines, qline)
		line := `{"messages": [` + qline + `]}`
		lines = append(lines, line)
	}

	stra := agent.NewStringArrayAgent(lines)
	model := "qwen2.5:14b" // "llama3.2-vision"
	chat := llm.NewChatWithPrompt(model, "You are a helpful assistant.  Please answer questions in one sentence.", nil)
	var pipe agent.AgentPipe
	pipe.AddAgent(stra)
	pipe.AddAgent(chat)
	defer pipe.Close()

	cnt := 0
	it, err := pipe.Execute(nil, nil)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	for data, err := range it {
		common.Assert(t, err == nil, "Expected nil, got %v", err)
		var output llm.ChatOutput
		err = json.Unmarshal(data, &output)
		common.Assert(t, err == nil, "Expected nil, got %v", err)

		t.Logf("Question: %s\n", qlines[cnt])
		t.Logf("Answer: %s\n", output.Response.Message.Content)
		cnt++
	}
}
