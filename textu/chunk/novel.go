package chunk

import (
	"bufio"
	"bytes"
	"io"
	"iter"
)

type NovelChunker struct {
	// input
	r    io.Reader
	scan *bufio.Scanner

	// state
	num1       int32
	num2       int32
	offset     int32
	path       []string
	pathString string
	buf        bytes.Buffer
	emptyLine  int
	done       bool
}

// Novel chunker chunks a novel (text, for example, downloaded from Project Gutenberg)
// into chunks.  It assumes that a new chunk starts after an empty line.  Chapters are
// assumed to be separated by multiple empty lines.
func NewNovelChunker(r io.Reader) (Chunker, error) {
	return &NovelChunker{
		r:    r,
		scan: bufio.NewScanner(r),
	}, nil
}

func (c *NovelChunker) handleLine(yield func(Chunk) bool) {
	// reset local chunk number if we have multiple empty lines
	c.num2++ // increment local chunk number

	// Prepare the chunk, reset the buffer
	chunk := Chunk{
		Num1: c.num1,
		Num2: c.num2,
		Text: c.buf.String(),
	}
	c.buf.Reset()

	// reset chunk number if we have multiple empty lines
	// note that we do this reset AFTER we prepare the current
	// chunk
	if c.emptyLine > 1 {
		c.num1++
		c.num2 = 0
	}
	c.emptyLine = 0

	// skip empty paragraph, only yield non-empty paragraph
	if chunk.Text != "" {
		if !yield(chunk) {
			c.done = true
		}
	}
}

func (c *NovelChunker) Chunk() iter.Seq[Chunk] {
	return func(yield func(Chunk) bool) {
		for c.scan.Scan() {
			line := c.scan.Text()
			if line != "" {
				if c.emptyLine > 0 {
					c.handleLine(yield)
					// handleLine will reset the buffer
				}
				// accumulate the buffer
				c.buf.WriteRune(' ')
				c.buf.WriteString(line)
			} else {
				c.emptyLine++
			}

			if c.done {
				break
			}
		}
		c.handleLine(yield)
	}
}
