package chunk

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/matrixorigin/monlp/common"
)

func TestParagraphChunker(t *testing.T) {
	input := `


Paragraph 1.1

Paragraph 1.2



Paragraph 2.1

Paragraph 2.2
Multiple line
Does not break paragraph

Paragraph 2.3


Paragraph 3.1

Paragraph 3.2
`

	// create a reader from input
	r := strings.NewReader(input)

	// create a new paragraph chunker
	chunker, err := NewNovelChunker(r)
	if err != nil {
		t.Errorf("Failed to create paragraph chunker: %v", err)
		return
	}

	// test the chunker
	nchunk := 0
	for chunk := range chunker.Chunk() {
		nchunk++
		prefix := fmt.Sprintf(" Paragraph %d.%d", chunk.Num1, chunk.Num2)
		if !strings.HasPrefix(chunk.Text, prefix) {
			t.Errorf("Expected prefix %q, got %q", prefix, chunk.Text)
		}
		t.Logf("Chunk: %d.%d: %s", chunk.Num1, chunk.Num2, chunk.Text)
	}

	if nchunk != 7 {
		t.Errorf("Expected 7 chunks, got %d", nchunk)
	}
}

func getFileReader(t *testing.T, filename string) *os.File {
	_, fn, _, _ := runtime.Caller(0)
	dir := path.Dir(fn)
	fpath := path.Join(dir, "..", "data", filename)
	f, err := os.Open(fpath)
	if err != nil {
		t.Errorf("Failed to open file %s: %v", filename, err)
		return nil
	}
	return f
}

func testReaderChunker(t *testing.T, fn string) {
	f, err := common.OpenFileForTest("data", fn)
	if err != nil {
		t.Errorf("Failed to read shakespeare.txt, err %v", err)
		return
	}
	defer f.Close()

	// create a new paragraph chunker
	chunker, err := NewNovelChunker(f)
	if err != nil {
		t.Errorf("Failed to create paragraph chunker: %v", err)
		return
	}

	// test the chunker
	nchunk := 0
	var num1, num2 int32
	for chunk := range chunker.Chunk() {
		nchunk++
		num1 = chunk.Num1
		num2 = chunk.Num2
		// t.Logf("Chunk: %d.%d: %s", chunk.Num1, chunk.Num2, chunk.Text)
	}

	t.Logf("Total %d chunks, %d.%d", nchunk, num1, num2)
}

func TestNovelChunker(t *testing.T) {
	testReaderChunker(t, "t8.shakespeare.txt")
	testReaderChunker(t, "AnimalFarm.txt")
	testReaderChunker(t, "xyj.txt")
}
