package chunk

import (
	"iter"
)

type Chunk struct {
	Num1  int32  `json:"num1"`  // Sequence number of the chunk
	Num2  int32  `json:"num2"`  // Sequence number of the chunk under current path
	Path  string `json:"path"`  // Path of the chunk
	Title string `json:"title"` // Title of the chunk
	Text  string `json:"text"`  // Text of the chunk
}

type Chunker interface {
	Chunk() iter.Seq[Chunk]
}
