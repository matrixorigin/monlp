// Package dbagent implements a database agent that can read or write database.
package dbagent

import (
	"encoding/json"

	"github.com/matrixorigin/monlp/agent"
)

// DbAgent is the interface for agent, with an additional DB method to get database connection.
type DbAgent interface {
	agent.Agent
	DB() *MoDB
}

// Config is the configuration for dbagent.
type Config struct {
	Driver    string `json:"driver"`    // database driver
	ConnStr   string `json:"connstr"`   // connection string
	Table     string `json:"table"`     // table name
	QTemplate string `json:"qtemplate"` // query template
}

// DbQueryInput is the input for db query.
type DbQueryInput struct {
	// mode: exec or query (defuault "" means query)
	Mode string `json:"mode"`
	Data string `json:"data"`
}

// DbQueryOutput is the output for db query, for now it is a 2D string array.
// TODO: support typed columns
type DbQueryOutput struct {
	Data [][]string `json:"data"`
}

type dbQuery struct {
	agent.NilKVAgent
	agent.SimpleExecuteAgent
	conf Config
	db   *MoDB
}

func (c *dbQuery) DB() *MoDB {
	return c.db
}

func (c *dbQuery) Config(bs []byte) error {
	err := json.Unmarshal(bs, &c.conf)
	if err != nil {
		return err
	}

	c.db, err = OpenDB(c.conf.Driver, c.conf.ConnStr)
	if err != nil {
		return err
	}
	return nil
}

func (c *dbQuery) Close() error {
	return c.db.Close()
}

// NewDbQuery creates a new dbQuery agent.
func NewDbQuery() DbAgent {
	ca := &dbQuery{}
	ca.Self = ca
	return ca
}

func (c *dbQuery) ExecuteOne(input []byte, dict map[string]string, yield func([]byte, error) bool) error {
	if len(input) == 0 {
		return nil
	}

	// unmarshal input to DbQueryInput
	var dbQueryInput DbQueryInput
	err := json.Unmarshal(input, &dbQueryInput)
	if err != nil {
		return err
	}

	var rows [][]string
	if dbQueryInput.Mode == "exec" {
		err = c.db.Exec(dbQueryInput.Data)
	} else {
		rows, err = c.db.Query(dbQueryInput.Data)
	}
	if err != nil {
		return err
	}

	output := DbQueryOutput{Data: rows}
	bs, err := json.Marshal(output)
	if !yield(bs, err) {
		return agent.ErrYieldDone
	}
	return nil
}
