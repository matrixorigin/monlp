package chunker

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/matrixorigin/monlp/agent"
	"github.com/matrixorigin/monlp/textu/chunk"
)

type WikiChunkerInputData struct {
	Url string `json:"url"`
}

type WikiChunkerInput struct {
	Data WikiChunkerInputData `json:"data"`
}

type WikiChunkerOutput struct {
	Data [][]string
}

type wikiChunker struct {
	agent.NilKVAgent
	agent.NilConfigAgent
	agent.NilCloseAgent
	agent.SimpleExecuteAgent
	batchSize int
}

func NewWikiChunker(batchsz int) agent.Agent {
	ca := &wikiChunker{}
	ca.Self = ca
	ca.batchSize = batchsz
	return ca
}

func (c *wikiChunker) ExecuteOne(input []byte, dict map[string]string, yield func([]byte, error) bool) error {
	if len(input) == 0 {
		return nil
	}

	// unmarshal input to WikiChunkerInput
	var wikiChunkerInput WikiChunkerInput
	err := json.Unmarshal(input, &wikiChunkerInput)
	if err != nil {
		return err
	}

	// only handle file:// for now
	if len(wikiChunkerInput.Data.Url) < 7 || wikiChunkerInput.Data.Url[:7] != "file://" {
		return fmt.Errorf("invalid url: %s", wikiChunkerInput.Data.Url)
	}
	// Open the file
	file, err := os.Open(wikiChunkerInput.Data.Url[7:])
	if err != nil {
		return err
	}
	defer file.Close()

	// create chunker, chunking at page level.
	chunker, err := chunk.NewWikiChunker(file, false)

	npage := 0
	var output WikiChunkerOutput
	for chunk := range chunker.Chunk() {
		// if strings.ToLower(chunk.Title) == strings.ToLower(chunk.Path) {
		// this is a redirect based on case, let's ignore it.
		// continue
		// }

		output.Data = append(output.Data, []string{
			chunk.Title,
			strings.ToLower(chunk.Title),
			chunk.Path,
			chunk.Text})

		npage++
		if npage%c.batchSize == 0 {
			bs, err := json.Marshal(output)
			if err != nil {
				return err
			}
			if !yield(bs, nil) {
				return agent.ErrYieldDone
			}
			output.Data = nil
		}
	}

	// yield the last batch
	if len(output.Data) > 0 {
		bs, err := json.Marshal(output)
		if err != nil {
			return err
		}
		if !yield(bs, nil) {
			return agent.ErrYieldDone
		}
	}
	return nil
}
