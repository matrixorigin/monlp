package wikix

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/matrixorigin/monlp/agent"
)

// WikiX is the wiki explorer agent.   It take the same input/output
// as the llm chat agent.

type WikixTopic struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type WikixSubquery struct {
	Query  string `json:"query"`
	Answer string `json:"answer"`
}

type WikixLink struct {
	Title     string `json:"title"`
	Paragraph string `json:"paragraph"`
}

type WikixInfo struct {
	// Original User query
	UserQuery string `json:"user_query"`
	// Topics to explore
	Topics []WikixTopic `json:"topics"`
	// Subqueries to ask
	Subqueries []WikixSubquery `json:"subqueries"`

	// Current question
	Question string `json:"question"`
	// Current aritcle
	Article string `json:"article"`
	// Current links
	Links []WikixLink `json:"links"`
}

func (wi *WikixInfo) clear() {
	*wi = WikixInfo{}
}

type wikix struct {
	agent.NilCloseAgent
	agent.NilConfigAgent
	agent.SimpleExecuteAgent

	info      WikixInfo
	model     string
	sysPrompt string

	dir         string
	topicsTmpl  string
	linksTmpl   string
	subqTmpl    string
	summaryTmpl string
	finalTmpl   string
}

func NewWikiX(dir string, model, sysprompt string) (agent.Agent, error) {
	ca := &wikix{
		dir:       dir,
		model:     model,
		sysPrompt: sysprompt,
	}
	ca.Self = ca

	var err error

	// read a bunch of prompt templates
	ca.topicsTmpl, err = ca.loadTemp("wikix_topics.json")
	if err != nil {
		return nil, err
	}

	ca.linksTmpl, err = ca.loadTemp("wikix_links.json")
	if err != nil {
		return nil, err
	}

	ca.subqTmpl, err = ca.loadTemp("wikix_subq.json")
	if err != nil {
		return nil, err
	}

	ca.summaryTmpl, err = ca.loadTemp("wikix_summary.json")
	if err != nil {
		return nil, err
	}

	ca.finalTmpl, err = ca.loadTemp("wikix_final.json")
	if err != nil {
		return nil, err
	}

	return ca, err
}

func (c *wikix) loadTemp(fn string) (string, error) {
	fullFn := filepath.Join(c.dir, fn)
	bs, err := os.ReadFile(fullFn)
	return string(bs), err
}

func (c *wikix) SetValue(name string, value any) error {
	c.info.clear()
	switch name {
	case "model":
		c.model = value.(string)
	case "sysPrompt":
		c.sysPrompt = value.(string)
	case "userquery":
		c.info.UserQuery = value.(string)
	default:
		return fmt.Errorf("unknown name: %s", name)
	}
	return nil
}

func (c *wikix) ExecuteOne(input []byte, dict map[string]string, yield func([]byte, error) bool) error {
	return nil
}

func (c *wikix) runTopics() ([]WikixTopic, error) {
	return nil, nil
}
