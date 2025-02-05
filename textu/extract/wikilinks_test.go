package extract

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/matrixorigin/monlp/agent/dbagent"
	"github.com/matrixorigin/monlp/common"
)

type WikiPageData struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

func TestGenData(t *testing.T) {
	driver, connstr := common.DbConnInfoForTest()
	db, err := dbagent.OpenDB(driver, connstr)
	common.Assert(t, err == nil, "OpenDB failed: %v", err)

	wf, err := common.CreateFileForTest("data", "wiki2pages.txt")
	common.Assert(t, err == nil, "OpenDB failed: %v", err)

	var wp [2]WikiPageData

	wp[0].Title = "Michael Freedman"
	wp[0].Text, err = db.QueryVal("select content from wikipages where title = 'Michael Freedman'")
	common.Assert(t, err == nil, "QueryVal failed: %v", err)

	wp[1].Title = "SVD"
	wp[1].Text, err = db.QueryVal("select content from wikipages where title = 'SVD'")
	common.Assert(t, err == nil, "QueryVal failed: %v", err)

	bs, err := json.Marshal(wp)
	common.Assert(t, err == nil, "Marshal failed: %v", err)

	wf.Write(bs)
	wf.Close()
}

func TestLinkExtract(t *testing.T) {
	dataf, err := common.OpenFileForTest("data", "wiki2pages.txt")
	common.Assert(t, err == nil, "OpenFileForTest failed: %v", err)

	// read the data
	data, err := io.ReadAll(dataf)

	var wp2 [2]WikiPageData
	err = json.Unmarshal(data, &wp2)
	common.Assert(t, err == nil, "Unmarshal failed: %v", err)

	common.Assert(t, wp2[0].Title == "Michael Freedman", "Expected Michael Freedman, got %s", wp2[0].Title)
	common.Assert(t, wp2[1].Title == "SVD", "Expected SVD, got %s", wp2[1].Title)

	var ex WikiLinkExtractor

	nlink := 0
	for _, wp := range wp2 {
		for link := range ex.Extract(wp.Title, wp.Text) {
			nlink++
			t.Logf("%s Link: offset: %d, value %s.%s, text: %s", wp.Title, link.Offset, link.Value, link.Value2, link.Text)
		}
	}
	t.Logf("Total links: %d", nlink)
}
