package oursql

import (
	"bytes"
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
		TestConnParam.DbName = "oursql_test"
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
		_, err = conn.Execute("INSERT INTO test VALUES(" + strconv.Itoa(i) + ",'" + strconv.Itoa(i) + "')")
		if err != nil {
			t.Fatal(err)
		}
	}
}

func Test_QuerySet(t *testing.T) {
	conn, err := Connect(TestConnParam)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	var res *DataSet

	res, err = conn.QuerySet("SELECT * FROM test ORDER BY id ASC", 0, true)
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Rows) != 10 {
		t.Fatalf("len(res.Rows) != 10, %d", len(res.Rows))
	}

	for i := 0; i < 10; i++ {
		v := []byte(strconv.Itoa(i))

		if !bytes.Equal(res.Rows[i][0].Inner, v) {
			t.Fatalf("id not match: %s", v)
		}

		if !bytes.Equal(res.Rows[i][1].Inner, v) {
			t.Fatalf("value not match: %s", v)
		}
	}
}

func Test_QueryReader(t *testing.T) {
	conn, err := Connect(TestConnParam)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	var res *DataReader

	res, err = conn.QueryReader("SELECT * FROM test ORDER BY id ASC", 0, true)
	if err != nil {
		t.Fatal(err)
	}

	i := 0
	for {
		row, err1 := res.FetchNext()
		if err1 != nil {
			t.Fatal(err1)
		}

		if row == nil {
			break
		}

		v := []byte(strconv.Itoa(i))

		if !bytes.Equal(row[0].Inner, v) {
			t.Fatalf("id not match: %s", v)
		}

		if !bytes.Equal(row[1].Inner, v) {
			t.Fatalf("value not match: %s", v)
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

	_, err = conn.Prepare("INSERT INTO test VALUES(?, ?)")
	if err != nil {
		t.Fatal(err)
	}
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
