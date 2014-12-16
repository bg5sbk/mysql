package driver

import (
	"database/sql"
	"encoding/json"
	"github.com/funny/unitest"
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
	unitest.NotError(t, err)

	err = conn.Close()
	unitest.NotError(t, err)
}

func Test_Execute(t *testing.T) {
	param := TestConnEnv
	param.DbName = "mysql"
	name, _ := json.Marshal(param)

	conn, err := sql.Open("mysql", string(name))
	unitest.NotError(t, err)
	defer conn.Close()

	_, err = conn.Exec("CREATE DATABASE " + TestConnEnv.DbName)
	unitest.NotError(t, err)

	_, err = conn.Exec("USE " + TestConnEnv.DbName)
	unitest.NotError(t, err)

	_, err = conn.Exec(`CREATE TABLE test (
		id INT PRIMARY KEY,
		value VARCHAR(10)
	)`)
	unitest.NotError(t, err)

	for i := 0; i < 10; i++ {
		res, err := conn.Exec("INSERT INTO test VALUES(" + strconv.Itoa(i) + ",'" + strconv.Itoa(i) + "')")
		unitest.NotError(t, err)
		num, _ := res.RowsAffected()
		unitest.Pass(t, num == 1)
	}
}

func Test_Query(t *testing.T) {
	conn, err := sql.Open("mysql", TestConnParam)
	unitest.NotError(t, err)
	defer conn.Close()

	var res *sql.Rows

	res, err = conn.Query("SELECT * FROM test ORDER BY id ASC")
	unitest.NotError(t, err)
	defer res.Close()

	i := 0
	for res.Next() {
		var (
			id    int64
			value string
		)

		err := res.Scan(&id, &value)
		unitest.NotError(t, err)

		unitest.Pass(t, id == int64(i))
		unitest.Pass(t, value == strconv.Itoa(i))
		i++
	}

	unitest.Pass(t, i == 10)
}

func Test_Prepare(t *testing.T) {
	conn, err := sql.Open("mysql", TestConnParam)
	unitest.NotError(t, err)
	defer conn.Close()

	var (
		stmt   *sql.Stmt
		res    sql.Result
		reader *sql.Rows
	)

	stmt, err = conn.Prepare("SELECT * FROM test ORDER BY id ASC")
	unitest.NotError(t, err)

	reader, err = stmt.Query()
	unitest.NotError(t, err)

	i := 0
	for reader.Next() {
		var (
			id    int64
			value string
		)

		err := reader.Scan(&id, &value)
		unitest.NotError(t, err)

		unitest.Pass(t, id == int64(i))
		unitest.Pass(t, value == strconv.Itoa(i))
		i++
	}

	unitest.Pass(t, i == 10)

	stmt.Close()

	stmt, err = conn.Prepare("INSERT INTO test VALUES(?, ?)")
	unitest.NotError(t, err)

	res, err = stmt.Exec(10, "10")
	unitest.NotError(t, err)
	num, _ := res.RowsAffected()
	unitest.Pass(t, num == 1)

	stmt.Close()
}

func Test_Clean(t *testing.T) {
	conn, err := sql.Open("mysql", TestConnParam)
	unitest.NotError(t, err)
	defer conn.Close()

	_, err = conn.Exec("DROP DATABASE " + TestConnEnv.DbName)
	unitest.NotError(t, err)
}
