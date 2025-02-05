package chunk

import (
	"bytes"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"strings"

	"github.com/dustin/go-wikiparse"
)

type WikiChunker struct {
	// input
	r              io.Reader
	p              wikiparse.Parser
	chunkParagraph bool

	// state
	num1 int32
	num2 int32
	done bool
}

func NewWikiChunker(r io.Reader, chunkParagraph bool) (Chunker, error) {
	var err error
	c := &WikiChunker{
		r:              r,
		chunkParagraph: chunkParagraph,
	}

	c.p, err = wikiparse.NewParser(r)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *WikiChunker) handlePage(page *wikiparse.Page, yield func(Chunk) bool) {
	c.num1++
	c.num2 = 0

	var text string
	if page.Redir.Title == "" {
		if len(page.Revisions) != 1 {
			msg := fmt.Sprintf("Warning: more than one revision (%d) for page %s", len(page.Revisions), page.Title)
			slog.Info(msg)
			text = "MULTI REVISIONS: " + page.Title
		}
		for _, rev := range page.Revisions {
			// get first revision text
			text = rev.Text
			break
		}
	} else {
		// redirect page, just yield the redirect
		if !yield(Chunk{
			Num1:  c.num1,
			Num2:  c.num2,
			Path:  page.Redir.Title,
			Title: page.Title,
		}) {
			c.done = true
		}
		return
	}

	// real page, let's chunk it
	emptyLine := 0
	if c.chunkParagraph {
		// break text into lines
		var buf bytes.Buffer
		lines := strings.Split(text, "\n")
		for _, line := range lines {
			if line != "" {
				if emptyLine > 0 {
					c.num2++
					if !yield(Chunk{
						Num1:  c.num1,
						Num2:  c.num2,
						Title: page.Title,
						Text:  buf.String(),
					}) {
						c.done = true
						return
					}
					buf.Reset()
					emptyLine = 0
				}
				buf.WriteString(line)
			} else {
				emptyLine++
			}
		}
	} else {
		// TODO: handle a wiki page
		if !yield(Chunk{
			Num1:  c.num1,
			Num2:  c.num2,
			Path:  page.Redir.Title,
			Title: page.Title,
			Text:  text,
		}) {
			c.done = true
		}
	}
}

func (c *WikiChunker) Chunk() iter.Seq[Chunk] {
	return func(yield func(Chunk) bool) {
		for page, err := c.p.Next(); err == nil; page, err = c.p.Next() {
			c.handlePage(page, yield)
			// fmt.Println(page.Title)
			if c.done {
				break
			}
		}
	}
}
