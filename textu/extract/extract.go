package extract

type Extract struct {
	Offset  int    `json:"offset"`  // Offset of the extracted info
	Value   string `json:"value"`   // Value of the extracted info
	Value2  string `json:"value2"`  // Value2 of the extracted info
	Text    string `json:"text"`    // Text of the extracted info
	Context string `json:"context"` // Context of the extracted info
}
