package mysql

import (
	"github.com/funny/utest"
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
	TestConnParam.User = env("TEST_MYSQL_USER", "root")
	TestConnParam.Pass = env("TEST_MYSQL_PASS", "")
}

func Test_Connect(t *testing.T) {
	param := TestConnParam
	param.DbName = "mysql"

	conn, err := Connect(param)
	utest.IsNilNow(t, err)

	conn.Close()
	utest.Assert(t, conn.IsClosed())
}

func Test_Execute(t *testing.T) {
	param := TestConnParam
	param.DbName = "mysql"

	conn, err := Connect(param)
	utest.IsNilNow(t, err)
	defer conn.Close()

	_, err = conn.Execute("CREATE DATABASE " + TestConnParam.DbName)
	utest.IsNilNow(t, err)

	_, err = conn.Execute("USE " + TestConnParam.DbName)
	utest.IsNilNow(t, err)

	_, err = conn.Execute(`CREATE TABLE test (
		id INT PRIMARY KEY,
		value VARCHAR(10)
	)`)
	utest.IsNilNow(t, err)

	for i := 0; i < 10; i++ {
		res, err := conn.Execute("INSERT INTO test VALUES(" + strconv.Itoa(i) + ",'" + strconv.Itoa(i) + "')")
		utest.IsNilNow(t, err)
		utest.EqualNow(t, res.RowsAffected(), 1)
	}
}

func Test_QueryTable(t *testing.T) {
	conn, err := Connect(TestConnParam)
	utest.IsNilNow(t, err)
	defer conn.Close()

	var res DataTable

	res, err = conn.QueryTable("SELECT * FROM test ORDER BY id ASC")
	utest.IsNilNow(t, err)

	rows := res.Rows()
	utest.EqualNow(t, len(rows), 10)

	for i := 0; i < 10; i++ {
		utest.EqualNow(t, rows[i][0].Int64(), int64(i))
		utest.EqualNow(t, rows[i][1].String(), strconv.Itoa(i))
	}
}

func Test_QueryReader(t *testing.T) {
	conn, err := Connect(TestConnParam)
	utest.IsNilNow(t, err)
	defer conn.Close()

	var res DataReader

	res, err = conn.QueryReader("SELECT * FROM test ORDER BY id ASC")
	utest.IsNilNow(t, err)
	defer res.Close()

	i := 0
	for {
		row, err1 := res.FetchNext()
		utest.IsNilNow(t, err1)

		if row == nil {
			break
		}

		utest.EqualNow(t, row[0].Int64(), int64(i))
		utest.EqualNow(t, row[1].String(), strconv.Itoa(i))
		i++
	}

	utest.EqualNow(t, i, 10)
}

func Test_Prepare(t *testing.T) {
	conn, err := Connect(TestConnParam)
	utest.IsNilNow(t, err)
	defer conn.Close()

	var (
		stmt   *Stmt
		res    Result
		table  DataTable
		reader DataReader
	)

	stmt, err = conn.Prepare("SELECT * FROM test ORDER BY id ASC")
	utest.IsNilNow(t, err)

	table, err = stmt.QueryTable()
	utest.IsNilNow(t, err)

	rows := table.Rows()
	utest.EqualNow(t, len(rows), 10)

	for i := 0; i < 10; i++ {
		utest.EqualNow(t, rows[i][0].Int64(), int64(i))
		utest.EqualNow(t, rows[i][1].String(), strconv.Itoa(i))
	}

	reader, err = stmt.QueryReader()
	utest.IsNilNow(t, err)

	i := 0
	for {
		row, err1 := reader.FetchNext()
		utest.IsNilNow(t, err1)

		if row == nil {
			break
		}

		utest.EqualNow(t, row[0].Int64(), int64(i))
		utest.EqualNow(t, row[1].String(), strconv.Itoa(i))

		i++
	}

	utest.EqualNow(t, i, 10)

	stmt.Close()

	stmt, err = conn.Prepare("INSERT INTO test VALUES(?, ?)")
	utest.IsNilNow(t, err)

	stmt.BindInt(10)
	stmt.BindText("10")

	res, err = stmt.Execute()
	utest.IsNilNow(t, err)
	utest.EqualNow(t, res.RowsAffected(), 1)

	stmt.Close()
}

func Test_Null(t *testing.T) {
	conn, err := Connect(TestConnParam)
	utest.IsNilNow(t, err)
	defer conn.Close()

	var res DataTable

	res, err = conn.QueryTable("SELECT SUM(value) FROM test")
	utest.IsNilNow(t, err)

	rows1 := res.Rows()
	utest.EqualNow(t, len(rows1), 1)
	utest.Assert(t, !rows1[0][0].IsNull())

	res, err = conn.QueryTable("SELECT SUM(value) FROM test WHERE id > 8888")
	utest.IsNilNow(t, err)

	rows2 := res.Rows()
	utest.EqualNow(t, len(rows2), 1)
	utest.Assert(t, rows2[0][0].IsNull())
}

func Test_Clean(t *testing.T) {
	conn, err := Connect(TestConnParam)
	utest.IsNilNow(t, err)
	defer conn.Close()

	_, err = conn.Execute("DROP DATABASE " + TestConnParam.DbName)
	utest.IsNilNow(t, err)
}
