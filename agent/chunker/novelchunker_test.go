package chunker

import (
	"testing"

	"github.com/matrixorigin/monlp/agent"
	"github.com/matrixorigin/monlp/common"
)

func TestNovelChunker(t *testing.T) {
	// test data
	book1 := "file://" + common.ProjectPath("data", "t8.shakespeare.txt")
	book2 := "file://" + common.ProjectPath("data", "AnimalFarm.txt")
	book3 := "file://" + common.ProjectPath("data", "xyj.txt")

	nca := NewNovelChunker()
	// optional
	nca.Config(nil)

	stra := agent.NewStringArrayAgent([]string{
		`{"data": {"url": "` + book1 + `"}}`,
		`{"data": {"url": "http://www.google.com"}}`,
		`{"data": {"url": "` + book2 + `"}}`,
		`{"data": {"url": "file://DoesNotExist"}}`,
		`{"data": {"url": "` + book3 + `"}}`,
	})

	var pipe agent.AgentPipe
	defer pipe.Close()

	pipe.AddAgent(stra)
	pipe.AddAgent(nca)

	it, err := pipe.Execute(nil, nil)
	common.Assert(t, err == nil, "Expected nil, got %v", err)

	var errs []error
	var nchunk int
	var nbytes int
	for data, err := range it {
		if err != nil {
			errs = append(errs, err)
			continue
		}
		nchunk++
		nbytes += len(data)
	}

	common.Assert(t, nchunk == 3, "Expected 3 chunks, got %d", nchunk)
	common.Assert(t, len(errs) == 2, "Expected 2 errors, got %d", len(errs))
	for i, err := range errs {
		t.Logf("Error %d: %v", i, err)
	}
	t.Logf("Total %d chunks, %d bytes", nchunk, nbytes)
}

func TestNovelStrChunker(t *testing.T) {
	// test data
	book1 := "file://" + common.ProjectPath("data", "t8.shakespeare.txt")
	book2 := "file://" + common.ProjectPath("data", "AnimalFarm.txt")
	book3 := "file://" + common.ProjectPath("data", "xyj.txt")
	book4 := "file://" + common.ProjectPath("data", "红楼梦.txt")

	stra := agent.NewStringArrayAgent([]string{
		`{"data": {"url": "` + book1 + `"}}`,
		`{"data": {"url": "` + book2 + `"}}`,
		`{"data": {"url": "` + book3 + `"}}`,
		`{"data": {"url": "` + book4 + `"}}`,
	})

	nca := NewNovelChunker()
	nca.Config([]byte(`{"string_mode": true}`))

	var pipe agent.AgentPipe
	pipe.AddAgent(stra)
	pipe.AddAgent(nca)
	defer pipe.Close()

	it, err := pipe.Execute(nil, nil)
	common.Assert(t, err == nil, "Expected nil, got %v", err)

	var nchunk int
	var nbytes int
	for data, err := range it {
		common.Assert(t, err == nil, "Expected nil, got %v", err)
		nchunk++
		nbytes += len(data)
	}

	common.Assert(t, nchunk == 4, "Expected 4 chunks, got %d", nchunk)
	t.Logf("Total %d chunks, %d bytes", nchunk, nbytes)
}
