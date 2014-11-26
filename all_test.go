package mysql

import (
	"github.com/funny/ceshi"
	"os"
	"strconv"
	"testing"
)

var TestConnParam ConnectionParams

func env(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func init() {
	TestConnParam.Host = env("TEST_MYSQL_HOST", "127.0.0.1")
	TestConnParam.Port, _ = strconv.Atoi(env("TEST_MYSQL_PORT", "3306"))
	TestConnParam.DbName = env("TEST_MYSQL_DBNAME", "mysql_test")
	TestConnParam.Uname = env("TEST_MYSQL_UNAME", "root")
	TestConnParam.Pass = env("TEST_MYSQL_PASS", "")
}

func Test_Connect(t *testing.T) {
	param := TestConnParam
	param.DbName = "mysql"

	conn, err := Connect(param)
	ceshi.NotError(t, err)

	conn.Close()
	ceshi.Pass(t, conn.IsClosed())
}

func Test_Execute(t *testing.T) {
	param := TestConnParam
	param.DbName = "mysql"

	conn, err := Connect(param)
	ceshi.NotError(t, err)
	defer conn.Close()

	_, err = conn.Execute("CREATE DATABASE " + TestConnParam.DbName)
	ceshi.NotError(t, err)

	_, err = conn.Execute("USE " + TestConnParam.DbName)
	ceshi.NotError(t, err)

	_, err = conn.Execute(`CREATE TABLE test (
		id INT PRIMARY KEY,
		value VARCHAR(10)
	)`)
	ceshi.NotError(t, err)

	for i := 0; i < 10; i++ {
		res, err := conn.Execute("INSERT INTO test VALUES(" + strconv.Itoa(i) + ",'" + strconv.Itoa(i) + "')")
		ceshi.NotError(t, err)
		ceshi.Pass(t, res.RowsAffected() == 1)
	}
}

func Test_QueryTable(t *testing.T) {
	conn, err := Connect(TestConnParam)
	ceshi.NotError(t, err)
	defer conn.Close()

	var res DataTable

	res, err = conn.QueryTable("SELECT * FROM test ORDER BY id ASC")
	ceshi.NotError(t, err)

	rows := res.Rows()
	ceshi.Pass(t, len(rows) == 10)

	for i := 0; i < 10; i++ {
		ceshi.Pass(t, rows[i][0].Int() == int64(i))
		ceshi.Pass(t, rows[i][1].String() == strconv.Itoa(i))
	}
}

func Test_QueryReader(t *testing.T) {
	conn, err := Connect(TestConnParam)
	ceshi.NotError(t, err)
	defer conn.Close()

	var res DataReader

	res, err = conn.QueryReader("SELECT * FROM test ORDER BY id ASC")
	ceshi.NotError(t, err)
	defer res.Close()

	i := 0
	for {
		row, err1 := res.FetchNext()
		ceshi.NotError(t, err1)

		if row == nil {
			break
		}

		ceshi.Pass(t, row[0].Int() == int64(i))
		ceshi.Pass(t, row[1].String() == strconv.Itoa(i))
		i++
	}

	ceshi.Pass(t, i == 10)
}

func Test_Prepare(t *testing.T) {
	conn, err := Connect(TestConnParam)
	ceshi.NotError(t, err)
	defer conn.Close()

	var (
		stmt   *Stmt
		res    Result
		table  DataTable
		reader DataReader
	)

	stmt, err = conn.Prepare("SELECT * FROM test ORDER BY id ASC")
	ceshi.NotError(t, err)

	table, err = stmt.QueryTable()
	ceshi.NotError(t, err)

	rows := table.Rows()
	ceshi.Pass(t, len(rows) == 10)

	for i := 0; i < 10; i++ {
		ceshi.Pass(t, rows[i][0].Int() == int64(i))
		ceshi.Pass(t, rows[i][1].String() == strconv.Itoa(i))
	}

	reader, err = stmt.QueryReader()
	ceshi.NotError(t, err)

	i := 0
	for {
		row, err1 := reader.FetchNext()
		ceshi.NotError(t, err1)

		if row == nil {
			break
		}

		ceshi.Pass(t, row[0].Int() == int64(i))
		ceshi.Pass(t, row[1].String() == strconv.Itoa(i))

		i++
	}

	ceshi.Pass(t, i == 10)

	stmt.Close()

	stmt, err = conn.Prepare("INSERT INTO test VALUES(?, ?)")
	ceshi.NotError(t, err)

	stmt.BindInt(10)
	stmt.BindString("10")

	res, err = stmt.Execute()
	ceshi.NotError(t, err)
	ceshi.Pass(t, res.RowsAffected() == 1)

	stmt.Close()
}

func Test_Clean(t *testing.T) {
	conn, err := Connect(TestConnParam)
	ceshi.NotError(t, err)
	defer conn.Close()

	_, err = conn.Execute("DROP DATABASE " + TestConnParam.DbName)
	ceshi.NotError(t, err)
}
