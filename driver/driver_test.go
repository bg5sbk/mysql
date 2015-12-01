package driver

import (
	"database/sql"
	"encoding/json"
	"github.com/funny/utest"
	"os"
	"strconv"
	"testing"
)

var (
	TestConnEnv   connParams
	TestConnParam string
)

func env(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func init() {
	TestConnEnv.Host = env("TEST_MYSQL_HOST", "127.0.0.1")
	TestConnEnv.Port, _ = strconv.Atoi(env("TEST_MYSQL_PORT", "3306"))
	TestConnEnv.DbName = env("TEST_MYSQL_DBNAME", "mysql_test")
	TestConnEnv.Uname = env("TEST_MYSQL_UNAME", "root")
	TestConnEnv.Pass = env("TEST_MYSQL_PASS", "")

	name, err := json.Marshal(TestConnEnv)
	if err != nil {
		panic(err)
	}

	TestConnParam = string(name)

	sql.Register("mysql", MySqlDriver{})
}

func Test_Connect(t *testing.T) {
	conn, err := sql.Open("mysql", TestConnParam)
	utest.IsNilNow(t, err)

	err = conn.Close()
	utest.IsNilNow(t, err)
}

func Test_Execute(t *testing.T) {
	param := TestConnEnv
	param.DbName = "mysql"
	name, _ := json.Marshal(param)

	conn, err := sql.Open("mysql", string(name))
	utest.IsNilNow(t, err)
	defer conn.Close()

	_, err = conn.Exec("CREATE DATABASE " + TestConnEnv.DbName)
	utest.IsNilNow(t, err)

	_, err = conn.Exec("USE " + TestConnEnv.DbName)
	utest.IsNilNow(t, err)

	_, err = conn.Exec(`CREATE TABLE test (
		id INT PRIMARY KEY,
		value VARCHAR(10)
	)`)
	utest.IsNilNow(t, err)

	for i := 0; i < 10; i++ {
		res, err := conn.Exec("INSERT INTO test VALUES(" + strconv.Itoa(i) + ",'" + strconv.Itoa(i) + "')")
		utest.IsNilNow(t, err)
		num, _ := res.RowsAffected()
		utest.Equal(t, num, 1)
	}
}

func Test_Query(t *testing.T) {
	conn, err := sql.Open("mysql", TestConnParam)
	utest.IsNilNow(t, err)
	defer conn.Close()

	var res *sql.Rows

	res, err = conn.Query("SELECT * FROM test ORDER BY id ASC")
	utest.IsNilNow(t, err)
	defer res.Close()

	i := 0
	for res.Next() {
		var (
			id    int64
			value string
		)

		err := res.Scan(&id, &value)
		utest.IsNilNow(t, err)

		utest.Equal(t, id, int64(i))
		utest.Equal(t, value, strconv.Itoa(i))
		i++
	}

	utest.Equal(t, i, 10)
}

func Test_Prepare(t *testing.T) {
	conn, err := sql.Open("mysql", TestConnParam)
	utest.IsNilNow(t, err)
	defer conn.Close()

	var (
		stmt   *sql.Stmt
		res    sql.Result
		reader *sql.Rows
	)

	stmt, err = conn.Prepare("SELECT * FROM test ORDER BY id ASC")
	utest.IsNilNow(t, err)

	reader, err = stmt.Query()
	utest.IsNilNow(t, err)

	i := 0
	for reader.Next() {
		var (
			id    int64
			value string
		)

		err := reader.Scan(&id, &value)
		utest.IsNilNow(t, err)

		utest.Equal(t, id, int64(i))
		utest.Equal(t, value, strconv.Itoa(i))
		i++
	}

	utest.Equal(t, i, 10)

	stmt.Close()

	stmt, err = conn.Prepare("INSERT INTO test VALUES(?, ?)")
	utest.IsNilNow(t, err)

	res, err = stmt.Exec(10, "10")
	utest.IsNilNow(t, err)
	num, _ := res.RowsAffected()
	utest.Equal(t, num, 1)

	stmt.Close()
}

func Test_Clean(t *testing.T) {
	conn, err := sql.Open("mysql", TestConnParam)
	utest.IsNilNow(t, err)
	defer conn.Close()

	_, err = conn.Exec("DROP DATABASE " + TestConnEnv.DbName)
	utest.IsNilNow(t, err)
}
