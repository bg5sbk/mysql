package mysql

import (
	"os"
	"strconv"
	"testing"
)

var TestConnParam ConnectionParams

func init() {
	if os.Getenv("TEST_MYSQL_HOST") != "" {
		TestConnParam.Host = os.Getenv("TEST_MYSQL_HOST")
	} else {
		TestConnParam.Host = "127.0.0.1"
	}

	if os.Getenv("TEST_MYSQL_PORT") != "" {
		TestConnParam.Port, _ = strconv.Atoi(os.Getenv("TEST_MYSQL_PORT"))
	} else {
		TestConnParam.Port = 3306
	}

	if os.Getenv("TEST_MYSQL_DBNAME") != "" {
		TestConnParam.DbName = os.Getenv("TEST_MYSQL_DBNAME")
	} else {
		TestConnParam.DbName = "mysql_test"
	}

	if os.Getenv("TEST_MYSQL_UNAME") != "" {
		TestConnParam.Uname = os.Getenv("TEST_MYSQL_UNAME")
	} else {
		TestConnParam.Uname = "root"
	}

	if os.Getenv("TEST_MYSQL_PASS") != "" {
		TestConnParam.Pass = os.Getenv("TEST_MYSQL_PASS")
	}
}

func Test_Connect(t *testing.T) {
	param := TestConnParam
	param.DbName = "mysql"

	conn, err := Connect(param)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("conn.Id() == %d", conn.Id())

	conn.Close()

	if !conn.IsClosed() {
		t.Fatal()
	}
}

func Test_Execute(t *testing.T) {
	param := TestConnParam
	param.DbName = "mysql"

	conn, err := Connect(param)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	_, err = conn.Execute("CREATE DATABASE " + TestConnParam.DbName)
	if err != nil {
		t.Fatal(err)
	}

	_, err = conn.Execute("USE " + TestConnParam.DbName)
	if err != nil {
		t.Fatal(err)
	}

	_, err = conn.Execute(`CREATE TABLE test (
		id INT PRIMARY KEY,
		value VARCHAR(10)
	)`)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		res, err := conn.Execute("INSERT INTO test VALUES(" + strconv.Itoa(i) + ",'" + strconv.Itoa(i) + "')")

		if err != nil {
			t.Fatal(err)
		}

		if res.RowsAffected() != 1 {
			t.Fatal(res.RowsAffected() != 1)
		}
	}
}

func Test_QueryTable(t *testing.T) {
	conn, err := Connect(TestConnParam)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	var res DataTable

	res, err = conn.QueryTable("SELECT * FROM test ORDER BY id ASC")
	if err != nil {
		t.Fatal(err)
	}

	rows := res.Rows()

	if len(rows) != 10 {
		t.Fatalf("len(rows) != 10, %d", len(rows))
	}

	for i := 0; i < 10; i++ {
		if rows[i][0].Int() != int64(i) {
			t.Fatalf("id not match: %s", rows[i][0].String())
		}

		if rows[i][1].String() != strconv.Itoa(i) {
			t.Fatalf("id not match: %s", rows[i][1].String())
		}
	}
}

func Test_QueryReader(t *testing.T) {
	conn, err := Connect(TestConnParam)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	var res DataReader

	res, err = conn.QueryReader("SELECT * FROM test ORDER BY id ASC")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Close()

	i := 0
	for {
		row, err1 := res.FetchNext()
		if err1 != nil {
			t.Fatal(err1)
		}

		if row == nil {
			break
		}

		if row[0].Int() != int64(i) {
			t.Fatalf("id not match: %s", row[0].String())
		}

		if row[1].String() != strconv.Itoa(i) {
			t.Fatalf("id not match: %s", row[1].String())
		}

		i += 1
	}

	if i != 10 {
		t.Fatal("row number not match")
	}
}

func Test_Prepare(t *testing.T) {
	conn, err := Connect(TestConnParam)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	var (
		stmt   *Stmt
		res    Result
		table  DataTable
		reader DataReader
	)

	stmt, err = conn.Prepare("SELECT * FROM test ORDER BY id ASC")
	if err != nil {
		t.Fatal(err)
	}

	table, err = stmt.QueryTable()
	if err != nil {
		t.Fatal(err)
	}

	rows := table.Rows()

	if len(rows) != 10 {
		t.Fatalf("len(rows) != 10, %d", len(rows))
	}

	for i := 0; i < 10; i++ {
		if rows[i][0].Int() != int64(i) {
			t.Fatalf("id not match: %d", rows[i][0].Int())
		}

		if rows[i][1].String() != strconv.Itoa(i) {
			t.Fatalf("value not match: %s", rows[i][1].String())
		}
	}

	reader, err = stmt.QueryReader()

	i := 0
	for {
		row, err1 := reader.FetchNext()
		if err1 != nil {
			t.Fatal(err1)
		}

		if row == nil {
			break
		}

		if row[0].Int() != int64(i) {
			t.Fatalf("id not match: %s", row[0].String())
		}

		if row[1].String() != strconv.Itoa(i) {
			t.Fatalf("id not match: %s", row[1].String())
		}

		i += 1
	}

	if i != 10 {
		t.Fatal("row number not match")
	}

	stmt.Close()

	stmt, err = conn.Prepare("INSERT INTO test VALUES(?, ?)")
	if err != nil {
		t.Fatal(err)
	}
	stmt.BindInt(10)
	stmt.BindString("10")

	res, err = stmt.Execute()
	if err != nil {
		t.Fatal(err)
	}
	if res.RowsAffected() != 1 {
		t.Fatal(res.RowsAffected() != 1)
	}

	stmt.Close()
}

func Test_Clean(t *testing.T) {
	conn, err := Connect(TestConnParam)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	_, err = conn.Execute("DROP DATABASE " + TestConnParam.DbName)
	if err != nil {
		t.Fatal(err)
	}
}
