package dbagent

import (
	"encoding/json"
	"testing"

	"github.com/matrixorigin/monlp/agent"
	"github.com/matrixorigin/monlp/common"
)

func TestDbQuery(t *testing.T) {
	// Assume that monlp database is created.
	// connstr := ConnStr("localhost", "6001", "dump", "111", "monlp")
	// conf := Config{ConnStr: connstr, Table: "testt"}
	connstr := "monlp.db"
	conf := Config{
		Driver:  "sqlite",
		ConnStr: connstr,
		Table:   "testt",
	}
	// marshal conf to json
	config, err := json.Marshal(conf)
	common.PanicAssert(t, err == nil, "Expected nil, got %v", err)

	qa := NewDbQuery()
	err = qa.Config(config)
	common.PanicAssert(t, err == nil, "Expected nil, got %v", err)
	// later will use qa in pipe and it will be closed by pipe.
	// defer qa.Close()

	// test a few queries
	stra := agent.NewStringArrayAgent([]string{
		// sqlite does not have generate_series unless we load extension.
		// `{"data": "select * from generate_series(1, 3) t"}`,
		`{"mode": "exec", "data": "drop table if exists testt"}`,
		`{"mode": "exec", "data": "create table testt (a int, b text)"}`,
		`{"mode": "exec", "data": "insert into testt values (1, 'a'), (2, 'b')"}`,
		`{"data": "select * from testt"}`,
	})

	var pipe agent.AgentPipe
	pipe.AddAgent(stra)
	pipe.AddAgent(qa)
	defer pipe.Close()

	it, err := pipe.Execute(nil, nil)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	for data, err := range it {
		common.Assert(t, err == nil, "Expected nil, got %v", err)
		t.Logf("Query Result: %s", string(data))
	}

	// pipe a query and writer.
	var wpipe agent.AgentPipe
	wstra := agent.NewStringArrayAgent([]string{
		`{"data": "select * from testt"}`,
		`{"data": "select * from testt"}`,
		`{"data": "select * from testt"}`,
	})

	wqa := NewDbQuery()
	err = wqa.Config(config)
	common.PanicAssert(t, err == nil, "Expected nil, got %v", err)

	wa := NewDbWriter()
	wa.Config(config)
	wpipe.AddAgent(wstra)
	wpipe.AddAgent(wqa)
	wpipe.AddAgent(wa)

	nrows, err := wa.DB().QueryIVal("select count(*) from testt")
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	common.Assert(t, nrows == 2, "Expected 2, got %v", nrows)

	it, err = wpipe.Execute(nil, nil)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	// At this moment, dbwriter has not be executed.
	nrows, err = wa.DB().QueryIVal("select count(*) from testt")
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	common.Assert(t, nrows == 2, "Expected 2, got %v", nrows)

	var dbWriterOutput DbWriterOutput
	for data, err := range it {
		common.Assert(t, err == nil, "Expected nil, got %v", err)
		err = json.Unmarshal(data, &dbWriterOutput)
		common.Assert(t, err == nil, "Expected nil, got %v", err)
		t.Logf("Wrote %d rows.", dbWriterOutput.Data)
	}

	nrows, err = wa.DB().QueryIVal("select count(*) from testt")
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	common.Assert(t, nrows == 16, "Expected 16, got %v", nrows)
}

func TestDBTemplate(t *testing.T) {
	var db MoDB
	q, err := db.Template2Q("select * from testt where a = {{.a}} and b = '{{.b}}'", map[string]string{"a": "1", "b": "a"})
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	common.Assert(t, q == "select * from testt where a = 1 and b = 'a'", "Expected select * from testt where a = 1 and b = 'a', got %v", q)

	q, err = db.Template2Q("select * from testt where a = {{.a * 2}} and b = '{{.b}}'", map[string]string{"a": "1", "b": "a'"})
	// go template does not allow arithmetic operations.
	common.Assert(t, err != nil, "Expected error, got nil")
}
