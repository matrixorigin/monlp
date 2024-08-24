// Copyright 2024 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tokenizer

import (
	"bytes"
	"fmt"
	"iter"
	"strings"
	"unicode/utf8"
)

const (
	MAX_TOKEN_SIZE    = 23
	MARKER_PER_TOKENS = 100
)

type Token struct {
	TokenBytes [1 + MAX_TOKEN_SIZE]byte
	TokenPos   int64
}

type SimpleTokenizer struct {
	// input and output
	input []byte

	// the buffer to store the _token
	begin        int
	currTokenPos int
	nextMarker   int
	latinBuf     bytes.Buffer

	Done bool
	Err  error
}

func NewSimpleTokenizer(input []byte) *SimpleTokenizer {
	return &SimpleTokenizer{input: input, nextMarker: MARKER_PER_TOKENS}
}

func isBreakerRune(rune rune) bool {
	// See ASCII table
	return rune < 33 || rune == 127 || rune == utf8.RuneError
}

// Assume we already tested isBreakerRune.  Test if rune is 1 or 2 byte UTF-8
func isLatin(rune rune) bool {
	return rune < 0x7FF
}

type handler func(t *SimpleTokenizer, pos int, rune rune, yield func(Token) bool) handler

func beginToken(t *SimpleTokenizer, pos int, rune rune, yield func(Token) bool) handler {
	if isBreakerRune(rune) {
		t.begin = pos
		return breakerToken
	} else if isLatin(rune) {
		t.begin = pos
		return latinToken
	} else {
		t.begin = pos
		return cjkToken
	}
}

func breakerToken(t *SimpleTokenizer, pos int, rune rune, yield func(Token) bool) handler {
	if isBreakerRune(rune) {
		return breakerToken
	} else {
		// if the breaker is not a single space, we increase token count.
		if pos > t.begin+1 || t.input[t.begin] != ' ' {
			t.incrTokenCount(1, pos, yield)
		}
		if isLatin(rune) {
			t.begin = pos
			return latinToken
		} else {
			t.begin = pos
			return cjkToken
		}
	}
}

func latinToken(t *SimpleTokenizer, pos int, rune rune, yield func(Token) bool) handler {
	if isBreakerRune(rune) {
		t.outputLatin(pos, yield)
		t.begin = pos
		return breakerToken
	} else if isLatin(rune) {
		// noop
		return latinToken
	} else {
		t.outputLatin(pos, yield)
		t.begin = pos
		return cjkToken
	}
}

func cjkToken(t *SimpleTokenizer, pos int, rune rune, yield func(Token) bool) handler {
	if isBreakerRune(rune) {
		t.outputCJK(pos, yield)
		t.begin = pos
		return breakerToken
		/* } else if isLatin(rune) {
		t.outputCJK(pos, yield)
		t.begin = pos
		return latinToken
		*/
	} else {
		return cjkToken
	}
}

func (t *SimpleTokenizer) incrTokenCount(n int, pos int, yield func(Token) bool) {
	t.currTokenPos += n
	if t.currTokenPos >= t.nextMarker {
		mkStr := fmt.Sprintf("_MARKER_%d", t.nextMarker)
		token := Token{}
		token.TokenBytes[0] = byte(len(mkStr))
		copy(token.TokenBytes[1:], []byte(mkStr))
		token.TokenPos = int64(pos)
		if !yield(token) {
			t.Done = true
		}
		t.nextMarker += MARKER_PER_TOKENS
	}
}

// outputLatin outputs the latin token from t.begin to pos
// punctuation is removed, and all letters are converted to lower case.
// if the token is longer than MAX_TOKEN_SIZE, it will be truncated.
// single letter token is ignored, but do increase token position.
func (t *SimpleTokenizer) outputLatin(pos int, yield func(Token) bool) {
	t.latinBuf.Reset()

	ibuf := t.input[t.begin:pos]
	for i := 0; i < len(ibuf); i++ {
		if ibuf[i] > 127 {
			if t.latinBuf.Len() >= MAX_TOKEN_SIZE-1 {
				break
			} else {
				t.latinBuf.WriteByte(ibuf[i])
				t.latinBuf.WriteByte(ibuf[i+1])
				i += 1
			}
		} else {
			if t.latinBuf.Len() >= MAX_TOKEN_SIZE {
				break
			}
			if (ibuf[i] >= '0' && ibuf[i] <= '9') || (ibuf[i] >= 'a' && ibuf[i] <= 'z') || (ibuf[i] >= 'A' && ibuf[i] <= 'Z') {
				t.latinBuf.WriteByte(ibuf[i])
			}
		}
	}

	if t.latinBuf.Len() > 1 {
		ls := strings.ToLower(t.latinBuf.String())
		token := Token{}
		token.TokenBytes[0] = byte(len(ls))
		copy(token.TokenBytes[1:], []byte(ls))
		token.TokenPos = int64(t.currTokenPos)
		if !yield(token) {
			t.Done = true
			return
		}
	}
	t.incrTokenCount(1, pos, yield)
}

// outputCJK outputs the CJK token from t.begin to pos
// if token contains latin letter, we do not normalize like outputLatin
func (t *SimpleTokenizer) outputCJK(pos int, yield func(Token) bool) {
	ibuf := t.input[t.begin:pos]
	ia := 0
	_, ib := utf8.DecodeRune(ibuf)
	_, sz := utf8.DecodeRune(ibuf[ib:])
	ic := ib + sz
	_, sz = utf8.DecodeRune(ibuf[ic:])
	id := ic + sz

	for ia < id {
		token := Token{}
		token.TokenBytes[0] = byte(id - ia)
		copy(token.TokenBytes[1:], ibuf[ia:id])
		token.TokenPos = int64(t.currTokenPos)
		if !yield(token) {
			t.Done = true
			return
		}
		t.incrTokenCount(1, t.begin+ia, yield)
		if t.Done {
			return
		}

		ia = ib
		ib = ic
		ic = id
		_, sz = utf8.DecodeRune(ibuf[id:])
		id += sz
	}
}

func (t *SimpleTokenizer) Tokenize() iter.Seq[Token] {
	return func(yield func(Token) bool) {
		if len(t.input) == 0 {
			return
		}

		h := beginToken

		for pos, rune := range string(t.input) {
			if t.Done {
				return
			}

			h = h(t, pos, rune, yield)
			if h == nil {
				break
			}
		}

		// send a space to output last token
		h(t, len(t.input), ' ', yield)
	}
}
