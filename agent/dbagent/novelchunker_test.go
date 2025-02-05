package dbagent

import (
	"encoding/json"
	"testing"

	"github.com/matrixorigin/monlp/agent"
	"github.com/matrixorigin/monlp/agent/chunker"
	"github.com/matrixorigin/monlp/common"
)

func TestLoadNovelChunker(t *testing.T) {
	// open database, create table
	connstr := ConnStr("localhost", "6001", "dump", "111", "monlp")
	conf := Config{ConnStr: connstr, Table: "testnovel"}
	config, err := json.Marshal(conf)
	common.PanicAssert(t, err == nil, "Expected nil, got %v", err)

	qa := NewDbQuery()
	err = qa.Config(config)
	common.PanicAssert(t, err == nil, "Expected nil, got %v", err)

	stra := agent.NewStringArrayAgent([]string{
		`{"data": "drop table if exists testnovel"}`,
		`{"data": "create table testnovel (` +
			`url varchar(200) not null, ` +
			`num1 int not null, ` +
			`num2 int not null, ` +
			`path text, title text, content text, ` +
			`primary key (url, num1, num2))"}`,
	})

	var pipe agent.AgentPipe
	pipe.AddAgent(stra)
	pipe.AddAgent(qa)
	defer pipe.Close()

	it, err := pipe.Execute(nil, nil)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	for _, err := range it {
		common.Assert(t, err == nil, "Expected nil, got %v", err)
	}

	// pipeline, chunk novels and write to database.
	wstra := agent.NewStringArrayAgent(nil)

	ca := chunker.NewNovelChunker()
	ca.Config([]byte(`{"string_mode": true}`))

	wa := NewDbWriter()
	waconf := Config{
		ConnStr:   connstr,
		Table:     "testnovel",
		QTemplate: "insert into testnovel values ('{{.URL}}', ?, ?, ?, ?, ?)",
	}
	waconfig, err := json.Marshal(waconf)
	common.PanicAssert(t, err == nil, "Expected nil, got %v", err)
	wa.Config(waconfig)

	var wpipe agent.AgentPipe
	wpipe.AddAgent(wstra)
	wpipe.AddAgent(ca)
	wpipe.AddAgent(wa)
	defer wpipe.Close()

	// Insert animal farm chunks
	book := "file://" + common.ProjectPath("data", "AnimalFarm.txt")
	dict := make(map[string]string)
	dict["URL"] = "AnimalFarm"
	wstra.SetValue("values", []string{`{"data": {"url": "` + book + `"}}`})
	it, err = wpipe.Execute(nil, dict)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	for _, err := range it {
		common.Assert(t, err == nil, "Expected nil, got %v", err)
	}
	nrows, err := wa.DB().QueryIVal("select count(*) from testnovel where url = 'AnimalFarm'")
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	t.Logf("AnimalFarm.size: %v", nrows)

	book = "file://" + common.ProjectPath("data", "xyj.txt")
	dict["URL"] = "xyj"
	wstra.SetValue("values", []string{`{"data": {"url": "` + book + `"}}`})
	it, err = wpipe.Execute(nil, dict)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	for _, err := range it {
		common.Assert(t, err == nil, "Expected nil, got %v", err)
	}
	nrows, err = wa.DB().QueryIVal("select count(*) from testnovel where url = 'xyj'")
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	t.Logf("xyj.size: %v", nrows)

	// Insert t8.shakespear chunks
	book = "file://" + common.ProjectPath("data", "t8.shakespeare.txt")
	dict["URL"] = "shakespear"
	wstra.SetValue("values", []string{`{"data": {"url": "` + book + `"}}`})
	it, err = wpipe.Execute(nil, dict)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	for _, err := range it {
		common.Assert(t, err == nil, "Expected nil, got %v", err)
	}
	nrows, err = wa.DB().QueryIVal("select count(*) from testnovel where url = 'shakespear'")
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	t.Logf("shakespear.size: %v", nrows)

	// Insert HLM
	book = "file://" + common.ProjectPath("data", "红楼梦.txt")
	dict["URL"] = "HLM"
	wstra.SetValue("values", []string{`{"data": {"url": "` + book + `"}}`})
	// HLM is in GBK encoding
	ca.SetValue("encoding", "GBK")
	it, err = wpipe.Execute(nil, dict)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	for _, err := range it {
		common.Assert(t, err == nil, "Expected nil, got %v", err)
	}
	nrows, err = wa.DB().QueryIVal("select count(*) from testnovel where url = 'HLM'")
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	t.Logf("HLM.size: %v", nrows)
}
