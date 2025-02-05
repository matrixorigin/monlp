package main

import (
	"encoding/json"
	"strings"

	"github.com/abiosoft/ishell/v2"
	"github.com/matrixorigin/monlp/agent"
	"github.com/matrixorigin/monlp/agent/chunker"
	"github.com/matrixorigin/monlp/agent/dbagent"
	"github.com/matrixorigin/monlp/cmd/u"
	"github.com/matrixorigin/monlp/common"
	"github.com/matrixorigin/monlp/textu/chunk"
)

func wikiStatsCmd(c *ishell.Context) {
	// assume the wiki is already downloaded
	f, err := common.OpenFileForTest("data", "enwiki-latest-pages-articles-multistream.xml")
	common.Assert(nil, err == nil, "Failed to open enwiki-latest-pages-articles-multistream.xml, err %v", err)
	defer f.Close()

	chunker, err := chunk.NewWikiChunker(f, false)
	common.Assert(nil, err == nil, "Failed to create wiki chunker, err %v", err)

	nchunk := 0
	redirect := 0
	multirev := 0

	for chunk := range chunker.Chunk() {
		nchunk++
		if chunk.Path != "" {
			redirect++
		}
		if strings.HasPrefix(chunk.Text, "MULTI REV") {
			multirev++
		}
	}

	c.Printf("Total chunks: %d, # of redirects %d, multi rev %d\n", nchunk, redirect, multirev)
}

func wikiLoadCmd(c *ishell.Context) {
	// First run the table creation command
	connstr := u.ConnStr()
	conf := dbagent.Config{Driver: common.SqlDriver, ConnStr: connstr, Table: "wikipages"}
	config, err := json.Marshal(conf)
	common.PanicAssert(nil, err == nil, "Expected nil, got %v", err)

	qa := dbagent.NewDbQuery()
	err = qa.Config(config)
	common.PanicAssert(nil, err == nil, "Expected nil, got %v", err)

	iddef := u.IdDef("id")
	stra := agent.NewStringArrayAgent([]string{
		`{"mode": "exec", "data": "drop table if exists wikipages"}`,
		`{"mode": "exec", "data": "create table wikipages (` +
			iddef + `, ` +
			`title varchar(1000) not null, ` +
			`k varchar(1000) not null,` +
			`redirect varchar(1000), ` +
			`content text)"}`,
		`{"mode": "exec", "data": "create index wikipages_k_idx on wikipages (k)"}`,
		`{"mode": "exec", "data": "create index wikipages_title_idx on wikipages (title)"}`,
		`{"mode": "exec", "data": "drop table if exists wikilinks"}`,
		`{"mode": "exec", "data": "create table wikilinks (` +
			`tfrom varchar(1000) not null, ` +
			`idfrom int not null, ` +
			`tto varchar(1000) not null, ` +
			`offset int)"}`,
		`{"mode": "exec", "data": "create index wikilinks_kfrom_idx on wikilinks (tfrom, tto)"}`,
		`{"mode": "exec", "data": "create index wikilinks_kto_idx on wikilinks (tto, tfrom)"}`,
	})

	var pipe agent.AgentPipe
	pipe.AddAgent(stra)
	pipe.AddAgent(qa)
	defer pipe.Close()
	it, err := pipe.Execute(nil, nil)
	common.Assert(nil, err == nil, "Expected nil, got %v", err)

	for _, err := range it {
		common.Assert(nil, err == nil, "Expected nil, got %v", err)
	}

	// wpipe write databases.
	wikifile := "file://" + common.ProjectPath("data", "enwiki-latest-pages-articles-multistream.xml")
	wstra := agent.NewStringArrayAgent([]string{
		`{"data": {"url": "` + wikifile + `"}}`,
	})

	batchSz := 1000
	chunker := chunker.NewWikiChunker(batchSz)

	wa := dbagent.NewDbWriter()
	waconf := dbagent.Config{
		Driver:    common.SqlDriver,
		ConnStr:   connstr,
		Table:     "wikipages",
		QTemplate: "insert into wikipages(title, k, redirect, content) values (?, ?, ?, ?)",
	}
	waconfig, err := json.Marshal(waconf)
	common.PanicAssert(nil, err == nil, "Expected nil, got %v", err)
	wa.Config(waconfig)

	var wpipe agent.AgentPipe
	wpipe.AddAgent(wstra)
	wpipe.AddAgent(chunker)
	wpipe.AddAgent(wa)
	defer wpipe.Close()

	wit, err := wpipe.Execute(nil, nil)
	common.Assert(nil, err == nil, "Expected nil, got %v", err)

	nbatch := 0
	for _, err := range wit {
		common.Assert(nil, err == nil, "Expected nil, got %v", err)
		nbatch++
		c.Printf("Load wikipages nbatch %d, pages %d.\n", nbatch, nbatch*batchSz)
	}
}
