package dbagent

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/matrixorigin/monlp/agent"
)

// DbWriterInput is the rows for db writer.
type DbWriterInput struct {
	Data [][]string `json:"data"`
}

// DbWriterOutput is the output for db writer, number of rows.
type DbWriterOutput struct {
	Data int `json:"data"` // number of rows written
}

type dbWriter struct {
	agent.NilKVAgent
	agent.SimpleExecuteAgent
	conf Config
	db   *MoDB
	proj func([]string) []string
}

func (c *dbWriter) DB() *MoDB {
	return c.db
}

// NewDbWriter creates a new dbWriter agent.
func NewDbWriter() DbAgent {
	ca := &dbWriter{}
	ca.Self = ca
	return ca
}

func (c *dbWriter) Config(bs []byte) error {
	err := json.Unmarshal(bs, &c.conf)
	if err != nil {
		return err
	}

	c.db, err = OpenDB(c.conf.Driver, c.conf.ConnStr)
	if err != nil {
		return err
	}

	if c.conf.Table == "" {
		return fmt.Errorf("Table name is empty")
	}
	return nil
}

func (c *dbWriter) SetProj(proj func([]string) []string) {
	c.proj = proj
}

func (c *dbWriter) Close() error {
	return c.db.Close()
}

func (c *dbWriter) ExecuteOne(input []byte, dict map[string]string, yield func([]byte, error) bool) error {
	if len(input) == 0 {
		return nil
	}

	var dbWriterInput DbWriterInput
	err := json.Unmarshal(input, &dbWriterInput)
	if err != nil {
		return err
	}

	nRows := len(dbWriterInput.Data)
	if nRows == 0 {
		return nil
	}

	nCols := len(dbWriterInput.Data[0])
	if nCols == 0 {
		return fmt.Errorf("No columns")
	}

	var sql string
	if c.conf.QTemplate != "" {
		sql, err = c.db.Template2Q(c.conf.QTemplate, dict)
		if err != nil {
			return err
		}
	} else {
		sql = fmt.Sprintf("INSERT INTO %s VALUES (", c.conf.Table)
		for i := 0; i < nCols; i++ {
			if i != 0 {
				sql += ","
			}
			sql += "?"
		}
		sql += ")"
	}

	stmt, err := c.db.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	buf := make([]interface{}, nCols)

	// Insert all the rows in one transaction.
	// Should we limit batch size?
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}
	// txStmt will be closed by tx.Commit()
	txStmt := tx.Stmt(stmt)
	for _, row := range dbWriterInput.Data {
		if c.proj != nil {
			row = c.proj(row)
		}

		if len(row) != nCols {
			return fmt.Errorf("Row has %d columns, expected %d", len(row), nCols)
		}
		// copy row to buf, maybe I should quit and use gorm.
		for i, v := range row {
			buf[i] = v
		}

		slog.Debug("DbWritter write row", "row", row)

		_, err = txStmt.Exec(buf...)
		if err != nil {
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		return err
	}

	output := DbWriterOutput{Data: nRows}
	bs, err := json.Marshal(output)
	if err != nil {
		return err
	}
	if !yield(bs, nil) {
		return agent.ErrYieldDone
	}
	return nil
}
