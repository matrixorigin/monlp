package llm

//
// A simple chat agent with function calling.
//
import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/matrixorigin/monlp/agent"
	"github.com/ollama/ollama/api"
)

// const DefaultModel = "llama3.2-vision"
const DefaultModel = "qwen2.5:14b"

type ChatInput struct {
	Messages []api.Message `json:"messages"`
}

type ChatOutput struct {
	Response api.ChatResponse `json:"response"`
}

type ChatConfig struct {
	Model        string          `json:"model"`
	SystemPrompt api.Message     `json:"system_prompt"`
	Format       json.RawMessage `json:"format"`
	Tools        api.Tools       `json:"tools"`
}

type LLMFunctionCall func(api.ToolCallFunction) (string, error)

type chatter struct {
	agent.NilCloseAgent
	agent.SimpleExecuteAgent
	conf     ChatConfig
	req      api.ChatRequest
	toolcall LLMFunctionCall
}

func NewChatWithPrompt(model, sysprompt string, tc LLMFunctionCall) agent.Agent {
	ca := &chatter{}

	ca.conf.Model = model
	ca.conf.SystemPrompt = api.Message{Role: "system", Content: sysprompt}
	ca.toolcall = tc
	ca.Self = ca
	ca.buildRequest()
	return ca
}

func (c *chatter) Config(bs []byte) error {
	// unmarshal config
	if bs == nil {
		return nil
	}
	err := json.Unmarshal(bs, &c.conf)
	c.buildRequest()
	return err
}

func (c *chatter) SetValue(name string, value any) error {
	if name == "toolcall" {
		tc, ok := value.(func(api.ToolCallFunction) (string, error))
		if !ok {
			return fmt.Errorf("invalid toolcall function")
		}
		c.toolcall = tc
		return nil
	} else if name == "model" {
		c.conf.Model = value.(string)
		c.req.Model = c.conf.Model
		return nil
	} else if name == "format" {
		c.conf.Format = value.(json.RawMessage)
		c.req.Format = c.conf.Format
		return nil
	} else if name == "tools" {
		c.conf.Tools = value.(api.Tools)
		c.req.Tools = c.conf.Tools
		return nil
	}

	return fmt.Errorf("unknown name: %s", name)
}

func (c *chatter) buildRequest() {
	c.req = api.ChatRequest{
		Model:    c.conf.Model,
		Messages: nil,
		Stream:   new(bool), // stream response default to false
		Format:   c.conf.Format,
		Tools:    c.conf.Tools,
		Options: map[string]interface{}{
			"temperature": 0.0,
		},
	}
}

func (c *chatter) ExecuteOne(input []byte, dict map[string]string, yield func([]byte, error) bool) error {
	if len(input) == 0 {
		return nil
	}

	var chatInput ChatInput
	err := json.Unmarshal(input, &chatInput)
	if err != nil {
		return err
	}

	client, err := api.ClientFromEnvironment()
	if err != nil {
		return err
	}

	c.req.Messages = []api.Message{
		c.conf.SystemPrompt,
	}
	c.req.Messages = append(c.req.Messages, chatInput.Messages...)

	ctx := context.Background()

	var output ChatOutput

	for output.Response.Model == "" {
		err = client.Chat(ctx, &c.req, func(resp api.ChatResponse) error {
			if len(resp.Message.ToolCalls) > 0 {
				for _, tc := range resp.Message.ToolCalls {
					if c.toolcall == nil {
						return fmt.Errorf("toolcall function is not set")
					}
					res, err := c.toolcall(tc.Function)
					if err != nil {
						return err
					}

					c.req.Messages = append(c.req.Messages, api.Message{Role: "tool", Content: res})
				}
			} else {
				output.Response = resp
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	bs, err := json.Marshal(output)
	if !yield(bs, err) {
		return agent.ErrYieldDone
	}
	return nil
}
