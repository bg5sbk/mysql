package driver

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/funny/mysql"
)

// MySQL connection parameter.
type connParams struct {
	Host       string `json:"host"`     // MySQL server host name or IP address.
	Port       int    `json:"port"`     // MySQL server port number.
	Uname      string `json:"user"`     // MySQL user name.
	Pass       string `json:"passwd"`   // MySQL password.
	DbName     string `json:"database"` // database name.
	UnixSocket string `json:"unix"`     // Unix socket path when using unix socket connection.
	Charset    string `json:"charset"`  // Connection charactor set.
	Flags      string `json:"flags"`    // Client flags. See http://dev.mysql.com/doc/refman/5.6/en/mysql-real-connect.html
}

type MySqlDriver struct {
}

func (d MySqlDriver) Open(name string) (driver.Conn, error) {
	params := connParams{}
	if err := json.Unmarshal([]byte(name), &params); err != nil {
		return nil, err
	}

	conn, err := mysql.Connect(mysql.ConnectionParams{
		Host:       params.Host,
		Port:       params.Port,
		Uname:      params.Uname,
		Pass:       params.Pass,
		DbName:     params.DbName,
		UnixSocket: params.UnixSocket,
		Charset:    params.Charset,
	})
	if err != nil {
		return nil, err
	}

	return &MySqlConn{*conn}, nil
}

type MySqlConn struct {
	conn mysql.Connection
}

func (c *MySqlConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	if len(args) == 0 {
		result, err := c.conn.Execute(query)
		if err != nil {
			return nil, err
		}
		return &MySqlResult{result}, nil
	}

	stmt, err1 := c.conn.Prepare(query)
	if err1 != nil {
		return nil, err1
	}
	defer stmt.Close()

	for i := 0; i < len(args); i++ {
		stmt.Bind(args[i])
	}

	result, err2 := stmt.Execute()
	if err2 != nil {
		return nil, err2
	}

	return &MySqlResult{result}, nil
}

func (c *MySqlConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	if len(args) == 0 {
		rows, err := c.conn.QueryReader(query)
		if err != nil {
			return nil, err
		}
		return &MySqlRows{rows}, nil
	}

	stmt, err1 := c.conn.Prepare(query)
	if err1 != nil {
		return nil, err1
	}
	defer stmt.Close()

	for i := 0; i < len(args); i++ {
		stmt.Bind(args[i])
	}

	rows, err2 := stmt.QueryReader()
	if err2 != nil {
		return nil, err2
	}

	return &MySqlRows{rows}, nil
}

func (c *MySqlConn) Prepare(query string) (driver.Stmt, error) {
	stmt, err := c.conn.Prepare(query)
	if err != nil {
		return nil, err
	}
	return &MySqlStmt{*stmt}, nil
}

func (c *MySqlConn) Close() error {
	c.conn.Close()
	return nil
}

func (c *MySqlConn) Begin() (driver.Tx, error) {
	_, err := c.conn.Execute("begin")
	if err != nil {
		return nil, err
	}
	return &MySqlTx{&c.conn}, nil
}

type MySqlTx struct {
	conn *mysql.Connection
}

func (t *MySqlTx) Commit() error {
	_, err := t.conn.Execute("commit")
	return err
}

func (t *MySqlTx) Rollback() error {
	_, err := t.conn.Execute("rollback")
	return err
}

type MySqlStmt struct {
	stmt mysql.Stmt
}

func (s *MySqlStmt) Close() error {
	return s.stmt.Close()
}

func (s *MySqlStmt) NumInput() int {
	return s.stmt.NumInput()
}

func (s *MySqlStmt) Exec(args []driver.Value) (driver.Result, error) {
	s.stmt.CleanBind()
	for i := 0; i < len(args); i++ {
		s.stmt.Bind(args[i])
	}
	result, err := s.stmt.Execute()
	if err != nil {
		return nil, err
	}
	return &MySqlResult{result}, nil
}

func (s MySqlStmt) Query(args []driver.Value) (driver.Rows, error) {
	s.stmt.CleanBind()
	for i := 0; i < len(args); i++ {
		s.stmt.Bind(args[i])
	}
	rows, err := s.stmt.QueryReader()
	if err != nil {
		return nil, err
	}
	return &MySqlRows{rows}, nil
}

type MySqlResult struct {
	result mysql.Result
}

func (r *MySqlResult) LastInsertId() (int64, error) {
	return r.result.InsertId(), nil
}

func (r *MySqlResult) RowsAffected() (int64, error) {
	return r.result.RowsAffected(), nil
}

type MySqlRows struct {
	rows mysql.DataReader
}

func (r *MySqlRows) Columns() []string {
	fields := r.rows.Fields()
	cols := make([]string, len(fields))
	for i := 0; i < len(cols); i++ {
		cols[i] = fields[i].Name
	}
	return cols
}

func (r *MySqlRows) Close() error {
	r.rows.Close()
	// TODO: error returns
	return nil
}

func (r *MySqlRows) Next(dest []driver.Value) error {
	cols, err := r.rows.FetchNext()
	if err != nil {
		return err
	}
	for i := 0; i < len(cols); i++ {
		dest[i] = cols[i].Interface().(driver.Value)
	}
	return nil
}
