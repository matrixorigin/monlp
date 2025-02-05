package chunk

import (
	"testing"

	"github.com/matrixorigin/monlp/common"
)

func TestWikiChunker(t *testing.T) {
	f, err := common.OpenFileForTest("data", "enwiki-latest-pages-articles-multistream.xml")
	common.Assert(t, err == nil, "Failed to open enwiki-latest-pages-articles-multistream.xml, err %v", err)
	defer f.Close()

	chunker, err := NewWikiChunker(f, false)
	common.Assert(t, err == nil, "Failed to create wiki chunker, err %v", err)

	nchunk := 0
	for chunk := range chunker.Chunk() {
		nchunk++
		t.Logf("Chunk: %d: %s", nchunk, chunk.Title)
		if nchunk >= 100 {
			break
		}
	}
	common.Assert(t, nchunk == 100, "Expected 100 chunk, got %d", nchunk)
}

func TestWikiParaChunker(t *testing.T) {
	f, err := common.OpenFileForTest("data", "enwiki-latest-pages-articles-multistream.xml")
	common.Assert(t, err == nil, "Failed to open enwiki-latest-pages-articles-multistream.xml, err %v", err)
	defer f.Close()

	chunker, err := NewWikiChunker(f, true)
	common.Assert(t, err == nil, "Failed to create wiki chunker, err %v", err)

	nchunk := 0
	for chunk := range chunker.Chunk() {
		nchunk++
		t.Logf("Chunk: %d: %s", nchunk, chunk.Title)
		if nchunk >= 200 {
			break
		}
	}
	common.Assert(t, nchunk == 200, "Expected 100 chunk, got %d", nchunk)
}
