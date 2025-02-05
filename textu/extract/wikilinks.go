package extract

import (
	"iter"
	"strings"
)

type WikiLinkExtractor struct {
	title      string
	page       string
	lines      []string
	currLine   int
	currOffset int
}

func (w *WikiLinkExtractor) Extract(title, page string) iter.Seq[Extract] {
	// Extract wiki links, which are in the form [[link]].   We will not
	// handle links across lines for now, not sure if they are possible.
	w.title = title
	w.page = page
	w.lines = strings.Split(page, "\n")
	w.currLine = 0
	w.currOffset = 0

	return func(yield func(Extract) bool) {
		// find the next link, process line by line
		for w.currLine < len(w.lines) {
			line := w.lines[w.currLine]
			// handle the current line
			for w.currOffset < len(line) {
				if line[w.currOffset] == '[' &&
					w.currOffset+1 < len(line) &&
					line[w.currOffset+1] == '[' {
					// found [[, now w.currOffset is at the first [
					// Next to find the closing ]]
					start := w.currOffset + 2
					for w.currOffset < len(line) {
						if line[w.currOffset] == ']' &&
							w.currOffset+1 < len(line) &&
							line[w.currOffset+1] == ']' {
							// found ]].
							link := line[start:w.currOffset]
							// Now a link [[Page#section|display]] is

							parts := strings.Split(link, "|")
							subs := strings.Split(parts[0], "#")

							var ex Extract
							ex.Offset = start
							ex.Value = subs[0]
							if len(subs) > 1 {
								ex.Value2 = subs[1]
							}
							if len(parts) > 1 {
								ex.Text = parts[1]
							}
							ex.Context = line
							if !yield(ex) {
								return
							}
							w.currOffset += 2
							break
						} else {
							w.currOffset++
						}
					}
				} else {
					w.currOffset++
				}
			}
			// go to the next line
			w.currLine++
			w.currOffset = 0
		}
	}
}
