package mysql

import (
	"github.com/funny/unitest"
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
	unitest.NotError(t, err)

	conn.Close()
	unitest.Pass(t, conn.IsClosed())
}

func Test_Execute(t *testing.T) {
	param := TestConnParam
	param.DbName = "mysql"

	conn, err := Connect(param)
	unitest.NotError(t, err)
	defer conn.Close()

	_, err = conn.Execute("CREATE DATABASE " + TestConnParam.DbName)
	unitest.NotError(t, err)

	_, err = conn.Execute("USE " + TestConnParam.DbName)
	unitest.NotError(t, err)

	_, err = conn.Execute(`CREATE TABLE test (
		id INT PRIMARY KEY,
		value VARCHAR(10)
	)`)
	unitest.NotError(t, err)

	for i := 0; i < 10; i++ {
		res, err := conn.Execute("INSERT INTO test VALUES(" + strconv.Itoa(i) + ",'" + strconv.Itoa(i) + "')")
		unitest.NotError(t, err)
		unitest.Pass(t, res.RowsAffected() == 1)
	}
}

func Test_QueryTable(t *testing.T) {
	conn, err := Connect(TestConnParam)
	unitest.NotError(t, err)
	defer conn.Close()

	var res DataTable

	res, err = conn.QueryTable("SELECT * FROM test ORDER BY id ASC")
	unitest.NotError(t, err)

	rows := res.Rows()
	unitest.Pass(t, len(rows) == 10)

	for i := 0; i < 10; i++ {
		unitest.Pass(t, rows[i][0].Int() == int64(i))
		unitest.Pass(t, rows[i][1].String() == strconv.Itoa(i))
	}
}

func Test_QueryReader(t *testing.T) {
	conn, err := Connect(TestConnParam)
	unitest.NotError(t, err)
	defer conn.Close()

	var res DataReader

	res, err = conn.QueryReader("SELECT * FROM test ORDER BY id ASC")
	unitest.NotError(t, err)
	defer res.Close()

	i := 0
	for {
		row, err1 := res.FetchNext()
		unitest.NotError(t, err1)

		if row == nil {
			break
		}

		unitest.Pass(t, row[0].Int() == int64(i))
		unitest.Pass(t, row[1].String() == strconv.Itoa(i))
		i++
	}

	unitest.Pass(t, i == 10)
}

func Test_Prepare(t *testing.T) {
	conn, err := Connect(TestConnParam)
	unitest.NotError(t, err)
	defer conn.Close()

	var (
		stmt   *Stmt
		res    Result
		table  DataTable
		reader DataReader
	)

	stmt, err = conn.Prepare("SELECT * FROM test ORDER BY id ASC")
	unitest.NotError(t, err)

	table, err = stmt.QueryTable()
	unitest.NotError(t, err)

	rows := table.Rows()
	unitest.Pass(t, len(rows) == 10)

	for i := 0; i < 10; i++ {
		unitest.Pass(t, rows[i][0].Int() == int64(i))
		unitest.Pass(t, rows[i][1].String() == strconv.Itoa(i))
	}

	reader, err = stmt.QueryReader()
	unitest.NotError(t, err)

	i := 0
	for {
		row, err1 := reader.FetchNext()
		unitest.NotError(t, err1)

		if row == nil {
			break
		}

		unitest.Pass(t, row[0].Int() == int64(i))
		unitest.Pass(t, row[1].String() == strconv.Itoa(i))

		i++
	}

	unitest.Pass(t, i == 10)

	stmt.Close()

	stmt, err = conn.Prepare("INSERT INTO test VALUES(?, ?)")
	unitest.NotError(t, err)

	stmt.BindInt(10)
	stmt.BindString("10")

	res, err = stmt.Execute()
	unitest.NotError(t, err)
	unitest.Pass(t, res.RowsAffected() == 1)

	stmt.Close()
}

func Test_Clean(t *testing.T) {
	conn, err := Connect(TestConnParam)
	unitest.NotError(t, err)
	defer conn.Close()

	_, err = conn.Execute("DROP DATABASE " + TestConnParam.DbName)
	unitest.NotError(t, err)
}
