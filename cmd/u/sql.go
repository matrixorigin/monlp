package u

import (
	"path"
	"strings"

	"github.com/abiosoft/ishell/v2"
	"github.com/matrixorigin/monlp/agent/dbagent"
	"github.com/matrixorigin/monlp/common"
)

var (
	db *dbagent.MoDB
)

func ConnStr() string {
	var connstr string
	switch common.SqlDriver {
	case "sqlite", "sqlite3", "dslite", "dslite3":
		connstr = path.Join(common.WorkingDir, "monlp.db")
	default:
		connstr = dbagent.ConnStr("localhost", "6001", "dump", "111", "monlp")
	}
	return connstr
}

func IdDef(col string) string {
	switch common.SqlDriver {
	case "sqlite", "sqlite3", "dslite", "dslite3":
		return col + " integer primary key autoincrement"
	default:
		return col + " int auto_increment not null primary key"
	}
}

func openDB(args []string) error {
	if db != nil {
		return nil
	}

	var err error
	connstr := ConnStr()
	db, err = dbagent.OpenDB(common.SqlDriver, connstr)
	return err
}

func SqlCmd(c *ishell.Context) {
	if len(c.Args) == 0 {
		c.Println("No sql to execute")
		return
	}

	if err := openDB(nil); err != nil {
		c.Println(err)
		return
	}

	sql := strings.Join(c.Args, " ")
	res, err := db.QueryDump(sql)
	if err != nil {
		c.Println(err)
		return
	} else {
		c.Println(res)
	}
}
