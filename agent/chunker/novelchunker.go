package chunker

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/matrixorigin/monlp/agent"
	"github.com/matrixorigin/monlp/textu/chunk"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type NovelChunkerConfig struct {
	StringMode bool   `json:"string_mode"`
	Encoding   string `json:"encoding"`
}

// NovelChunker is a chunker for novels.
type NovelChunkerInputData struct {
	Url string `json:"url"`
}

type NovelChunkerInput struct {
	Data NovelChunkerInputData `json:"data"`
}

type NovelChunkerOutput struct {
	Data []chunk.Chunk `json:"data"`
}

type NovelChunkerStrOutput struct {
	Data [][]string `json:"data"`
}

type novelChunker struct {
	agent.NilCloseAgent
	agent.SimpleExecuteAgent
	conf NovelChunkerConfig
}

func NewNovelChunker() agent.Agent {
	ca := &novelChunker{}
	ca.Self = ca
	return ca
}

func (c *novelChunker) Config(bs []byte) error {
	// unmarshal config
	if bs == nil {
		return nil
	}
	err := json.Unmarshal(bs, &c.conf)
	return err
}

func (c *novelChunker) SetValue(name string, encoding any) error {
	if name == "encoding" {
		c.conf.Encoding = encoding.(string)
	}
	return fmt.Errorf("unknown name: %s", name)
}

func (c *novelChunker) ExecuteOne(data []byte, dict map[string]string, yield func([]byte, error) bool) error {
	var novelChunkerInput NovelChunkerInput
	err := json.Unmarshal(data, &novelChunkerInput)
	if err != nil {
		return err
	}

	// only handle file:// for now
	if len(novelChunkerInput.Data.Url) < 7 || novelChunkerInput.Data.Url[:7] != "file://" {
		return fmt.Errorf("invalid url: %s", novelChunkerInput.Data.Url)
	}

	// Open the file
	file, err := os.Open(novelChunkerInput.Data.Url[7:])
	if err != nil {
		return err
	}
	defer file.Close()

	var reader io.Reader = file

	if c.conf.Encoding != "" {
		if c.conf.Encoding == "GBK" {
			reader = transform.NewReader(file, simplifiedchinese.GBK.NewDecoder())
		} else {
			return fmt.Errorf("unknown encoding: %s", c.conf.Encoding)
		}
	}

	// Read the file, call chunker
	chunks, err := chunk.NewNovelChunker(reader)
	if err != nil {
		return err
	}

	// Marshal the output
	if c.conf.StringMode {
		var output NovelChunkerStrOutput
		for chunk := range chunks.Chunk() {
			row := make([]string, 5)
			row[0] = strconv.Itoa(int(chunk.Num1))
			row[1] = strconv.Itoa(int(chunk.Num2))
			row[2] = chunk.Path
			row[3] = chunk.Title
			row[4] = chunk.Text
			output.Data = append(output.Data, row)
		}
		bs, err := json.Marshal(output)
		if err != nil {
			return err
		}
		if !yield(bs, nil) {
			return agent.ErrYieldDone
		}
	} else {
		var output NovelChunkerOutput
		for chunk := range chunks.Chunk() {
			output.Data = append(output.Data, chunk)
		}
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
