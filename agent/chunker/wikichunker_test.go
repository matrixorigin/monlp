package chunker

import (
	"encoding/json"
	"testing"

	"github.com/matrixorigin/monlp/agent"
	"github.com/matrixorigin/monlp/common"
)

func TestWiki100Pages(t *testing.T) {
	// create a new wiki chunker, batch size 10
	wikifile := "file://" + common.ProjectPath("data", "enwiki-latest-pages-articles-multistream.xml")
	stra := agent.NewStringArrayAgent([]string{
		`{"data": {"url": "` + wikifile + `"}}`,
	})

	chunker := NewWikiChunker(10)
	var pipe agent.AgentPipe
	pipe.AddAgent(stra)
	pipe.AddAgent(chunker)
	defer pipe.Close()

	it, err := pipe.Execute(nil, nil)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	npage := 0
	nbatch := 0
	for data, err := range it {
		common.Assert(t, err == nil, "Expected nil, got %v", err)

		var pages WikiChunkerOutput
		err = json.Unmarshal(data, &pages)
		common.Assert(t, err == nil, "Expected nil, got %v", err)
		nbatch++
		npage += len(pages.Data)
		common.Assert(t, len(pages.Data) == 10, "Expected 10 pages, got %d", len(pages.Data))

		for idx, page := range pages.Data {
			if page[0] == "accessiblecomputing" {
				t.Logf("Batch %d.%d: %s, %s\n", nbatch, idx, page[0], page[1])
			}
		}
		if npage >= 100 {
			break
		}
	}

	common.Assert(t, npage == 100, "Expected 100 pages, got %d", npage)
}
