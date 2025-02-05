package llm

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/matrixorigin/monlp/agent"
	"github.com/matrixorigin/monlp/common"
	"github.com/ollama/ollama/api"
)

func TestSimpleChat(t *testing.T) {
	qs := ChatInput{
		Messages: []api.Message{
			{Role: "user", Content: "1+2="},
			{Role: "user", Content: "What is the color of the sky?"},
			{Role: "user", Content: "What is the capital of France?"},
		},
	}

	qss, err := json.Marshal(&qs)
	common.Assert(t, err == nil, "Expected nil, got %v", err)

	stra := agent.NewStringArrayAgent([]string{
		string(qss),
	})

	chat := NewChatWithPrompt("llama3.2-vision", "You are a helpful assistant.  You should answer each question in one word.", nil)

	var pipe agent.AgentPipe
	pipe.AddAgent(stra)
	pipe.AddAgent(chat)
	defer pipe.Close()

	it, err := pipe.Execute(nil, nil)
	common.Assert(t, err == nil, "Expected nil, got %v", err)

	for data, err := range it {
		common.Assert(t, err == nil, "Expected nil, got %v", err)
		t.Logf("Query Result: %s", string(data))
	}

	chat2 := NewChatWithPrompt("deepseek-r1:14b", "You are a helpful assistant.  You should answer each question in one word.", nil)
	var pipe2 agent.AgentPipe
	pipe2.AddAgent(stra)
	pipe2.AddAgent(chat2)
	defer pipe2.Close()

	it2, err := pipe2.Execute(nil, nil)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	for data, err := range it2 {
		common.Assert(t, err == nil, "Expected nil, got %v", err)
		t.Logf("Query Result: %s", string(data))
	}
}

func any2Float64(v any) (float64, error) {
	switch v := v.(type) {
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("unknown type: %T", v)
	}
}

func toolCall(tc api.ToolCallFunction) (string, error) {
	switch tc.Name {
	case "opAt":
		arg1, err := any2Float64(tc.Arguments["arg1"])
		if err != nil {
			return "", err
		}
		arg2, err := any2Float64(tc.Arguments["arg2"])
		if err != nil {
			return "", err
		}
		res := arg1 + arg2*2
		return fmt.Sprintf("%f", res), nil

	case "opHash":
		arg1, err := any2Float64(tc.Arguments["arg1"])
		if err != nil {
			return "", err
		}
		arg2, err := any2Float64(tc.Arguments["arg2"])
		if err != nil {
			return "", err
		}
		res := arg1 - arg2*2
		return fmt.Sprintf("%f", res), nil

	case "getWeather":
		return "almost sunny", nil
	}
	return "", fmt.Errorf("unknown tool: %s", tc.Name)
}

type QA struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

func TestJsonChat(t *testing.T) {
	// should use json format output.
	// XXX: as of jan 2025, if we use json format output ollama will
	// not call the tools.  so we disable json format output for now.
	// Workaround is to use a prompt that outputs json.
	jsonFormatOutput := false
	tools := []any{
		map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        "opAt",
				"description": "opAt(arg1, arg2) evaluates arg1 @ arg2 where arg1 and arg2 are numbers.",
				"parameters": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"arg1": map[string]any{
							"type":        "string",
							"description": "The first argument",
						},
						"arg2": map[string]any{
							"type":        "string",
							"description": "The second argument",
						},
					},
					"required": []string{"arg1", "arg2"},
				},
			},
		},
		map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        "opHash",
				"description": "opHash(arg1, arg2) evaluates arg1 # arg2 where arg1 and arg2 are numbers.",
				"parameters": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"arg1": map[string]any{
							"type":        "string",
							"description": "The first argument",
						},
						"arg2": map[string]any{
							"type":        "string",
							"description": "The second argument",
						},
					},
					"required": []string{"arg1", "arg2"},
				},
			},
		},
		map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        "getWeather",
				"description": "Get the weather in a given location",
				"parameters": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "The location that you want to get the weather for",
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}

	jsonTools, err := json.Marshal(tools)
	common.Assert(t, err == nil, "Expected nil, got %v", err)

	var toolsList api.Tools
	err = json.Unmarshal(jsonTools, &toolsList)
	common.Assert(t, err == nil, "Expected nil, got %v", err)

	qs := ChatInput{
		Messages: []api.Message{
			{Role: "user", Content: "Question: 1@2="},
		},
	}
	qs1, err := json.Marshal(&qs)
	common.Assert(t, err == nil, "Expected nil, got %v", err)

	qs = ChatInput{
		Messages: []api.Message{
			{Role: "user", Content: "Question: What is the color of the sky?"},
		},
	}
	qs2, err := json.Marshal(&qs)
	common.Assert(t, err == nil, "Expected nil, got %v", err)

	qs = ChatInput{
		Messages: []api.Message{
			{Role: "user", Content: "Question: opHash(2, 3) ="},
		},
	}
	qs3, err := json.Marshal(&qs)
	common.Assert(t, err == nil, "Expected nil, got %v", err)

	qs = ChatInput{
		Messages: []api.Message{
			{Role: "user", Content: "Question: What is the weather of San Jose?"},
		},
	}
	qs4, err := json.Marshal(&qs)
	common.Assert(t, err == nil, "Expected nil, got %v", err)

	stra := agent.NewStringArrayAgent([]string{
		string(qs1),
		string(qs2),
		string(qs3),
		string(qs4),
	})

	schema := map[string]any{
		"type": "array",
		"items": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"question": map[string]any{
					"type": "string",
				},
				"answer": map[string]any{
					"type": "string",
				},
			},
			"required": []any{"question", "answer"},
		},
	}
	jsonSchema, err := json.Marshal(schema)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	jsonRaw := json.RawMessage(jsonSchema)

	chat := NewChatWithPrompt(
		// As of Jan 2025, qwen2.5:14b is the best model for function calling.
		// esp, llama3.2-vision and deepseek-r1 do not enable tool calling in ollama.
		"qwen2.5:14b", // "llama3.1",
		`You are a helpful assistant.  You may use a list of tools when appropriate.  Your final answer should be valid json like the following.
		[
			"answer": "answer in one word or phrase"},
		]
		`,
		toolCall)
	if jsonFormatOutput {
		chat.SetValue("format", jsonRaw)
	}
	chat.SetValue("tools", toolsList)

	pipe := agent.AgentPipe{}
	pipe.AddAgent(stra)
	pipe.AddAgent(chat)
	defer pipe.Close()

	it, err := pipe.Execute(nil, nil)
	common.Assert(t, err == nil, "Expected nil, got %v", err)

	for data, err := range it {
		common.Assert(t, err == nil, "Expected nil, got %v", err)
		var out ChatOutput
		err = json.Unmarshal(data, &out)
		common.Assert(t, err == nil, "Expected nil, got %v", err)

		if jsonFormatOutput {
			var qas []QA
			err = json.Unmarshal([]byte(out.Response.Message.Content), &qas)
			common.Assert(t, err == nil, "Expected nil, got %v", err)
			for _, qa := range qas {
				t.Logf("Question: %s, Answer: %s", qa.Question, qa.Answer)
			}
		} else {
			t.Logf("Response: %s", out.Response.Message.Content)
		}
	}
}
