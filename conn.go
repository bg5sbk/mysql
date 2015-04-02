package mysql

/*
#include "cgo.h"
*/
import "C"
import (
	"unsafe"
)

type ClientFlag int64

const (
	// The client can handle expired passwords.
	// CF_CAN_HANDLE_EXPIRED_PASSWORDS = C.CAN_HANDLE_EXPIRED_PASSWORDS

	// Use compression protocol.
	CF_CLIENT_COMPRESS = ClientFlag(C.CLIENT_COMPRESS)

	// Return the number of found (matched) rows, not the number of changed rows.
	CF_CLIENT_FOUND_ROWS = ClientFlag(C.CLIENT_FOUND_ROWS)

	// Prevents the client library from installing a SIGPIPE signal handler.
	// This can be used to avoid conflicts with a handler that the application has already installed.
	CF_CLIENT_IGNORE_SIGPIPE = ClientFlag(C.CLIENT_IGNORE_SIGPIPE)

	// Permit spaces after function names. Makes all functions names reserved words.
	CF_CLIENT_IGNORE_SPACE = ClientFlag(C.CLIENT_IGNORE_SPACE)

	// Permit interactive_timeout seconds (instead of wait_timeout seconds) of inactivity before closing the connection.
	// The client's session wait_timeout variable is set to the value of the session interactive_timeout variable.
	CF_CLIENT_INTERACTIVE = ClientFlag(C.CLIENT_INTERACTIVE)

	// Enable LOAD DATA LOCAL handling.
	CF_CLIENT_LOCAL_FILES = ClientFlag(C.CLIENT_LOCAL_FILES)

	// Tell the server that the client can handle multiple result sets from multiple-statement executions or stored procedures.
	// This flag is automatically enabled if CLIENT_MULTI_STATEMENTS is enabled.
	CF_CLIENT_MULTI_RESULTS = ClientFlag(C.CLIENT_MULTI_RESULTS)

	// Tell the server that the client may send multiple statements in a single string (separated by “;”).
	// If this flag is not set, multiple-statement execution is disabled.
	CF_CLIENT_MULTI_STATEMENTS = ClientFlag(C.CLIENT_MULTI_STATEMENTS)

	// Do not permit the db_name.tbl_name.col_name syntax. This is for ODBC.
	// It causes the parser to generate an error if you use that syntax, which is useful for trapping bugs in some ODBC programs.
	CF_CLIENT_NO_SCHEMA = ClientFlag(C.CLIENT_NO_SCHEMA)

	// Remember options specified by calls to mysql_options().
	// Without this option, if mysql_real_connect() fails, you must repeat the mysql_options() calls before trying to connect again.
	// With this option, the mysql_options() calls need not be repeated.
	CF_CLIENT_REMEMBER_OPTIONS = ClientFlag(C.CLIENT_REMEMBER_OPTIONS)
)

func init() {
	// This needs to be called before threads begin to spawn.
	C.my_library_init()
}

// MySQL connection parameter.
type ConnectionParams struct {
	Host       string     `json:"host"`     // MySQL server host name or IP address.
	Port       int        `json:"port"`     // MySQL server port number.
	Uname      string     `json:"user"`     // MySQL user name.
	Pass       string     `json:"passwd"`   // MySQL password.
	DbName     string     `json:"database"` // database name.
	UnixSocket string     `json:"unix"`     // Unix socket path when using unix socket connection.
	Charset    string     `json:"charset"`  // Connection charactor set.
	Flags      ClientFlag `json:"-"`        // Client flags. See http://dev.mysql.com/doc/refman/5.6/en/mysql-real-connect.html
}

// MySQL connection.
type Connection struct {
	c      C.MYSQL
	closed bool
}

// Connect to MySQL server.
func Connect(params ConnectionParams) (conn *Connection, err error) {
	host := C.CString(params.Host)
	defer cfree(host)

	uname := C.CString(params.Uname)
	defer cfree(uname)

	pass := C.CString(params.Pass)
	defer cfree(pass)

	dbname := C.CString(params.DbName)
	defer cfree(dbname)

	unix_socket := C.CString(params.UnixSocket)
	defer cfree(unix_socket)

	charset := C.CString(params.Charset)
	defer cfree(charset)

	port := C.uint(params.Port)
	flags := C.ulong(params.Flags)

	conn = &Connection{}

	if C.my_open(&conn.c, host, uname, pass, dbname, port, unix_socket, charset, flags) != 0 {
		defer conn.Close()
		return nil, conn.lastError("")
	}

	return conn, nil
}

// Get current connection thread id.
func (conn *Connection) Id() int64 {
	return int64(C.my_thread_id(&conn.c))
}

// Close connection.
func (conn *Connection) Close() {
	if !conn.closed {
		C.my_close(&conn.c)
		conn.closed = true
	}
}

// Check connection is closed or not.
func (conn *Connection) IsClosed() bool {
	return conn.closed
}

// Toggles autocommit mode on/off
func (conn *Connection) Autocommit(mode bool) error {
	m := C.my_bool(0)
	if mode {
		m = C.my_bool(1)
	}
	if C.my_autocommit(&conn.c, m) != 0 {
		return conn.lastError("")
	}
	return nil
}

// Commit current transaction
func (conn *Connection) Commit() error {
	if C.my_commit(&conn.c) != 0 {
		return conn.lastError("")
	}
	return nil
}

// Rollback current transaction
func (conn *Connection) Rollback() error {
	if C.my_rollback(&conn.c) != 0 {
		return conn.lastError("")
	}
	return nil
}

// Escapes special characters in a string for use in an SQL statement,
// taking into account the current character set of the connection.
func (conn *Connection) Escape(from string) string {
	to := make([]byte, len(from)*2+1)
	length := C.my_real_escape_string(&conn.c, (*C.char)(bytePointer(to)), (*C.char)(stringPointer(from)), C.ulong(len(from)))
	return string(to[:length])
}

// Execute a non-query SQL.
func (conn *Connection) Execute(sql string) (Result, error) {
	res := &connResult{}

	err := conn.execute(sql, res, C.MY_MODE_NONE)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Execute a query and fill result into a DataTable.
func (conn *Connection) QueryTable(sql string) (DataTable, error) {
	res := &connDataTable{}

	err := conn.query(sql, &res.connQueryResult, C.MY_MODE_TABLE)
	if err != nil {
		return nil, err
	}
	defer res.close()

	err = res.fillRows(conn)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Execute a query and return the result reader. NOTE: Please remember close the reader.
func (conn *Connection) QueryReader(sql string) (DataReader, error) {
	res := &connDataReader{}

	err := conn.query(sql, &res.connQueryResult, C.MY_MODE_READER)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func cfree(str *C.char) {
	if str != nil {
		C.free(unsafe.Pointer(str))
	}
}

func (conn *Connection) execute(sql string, res *connResult, mode C.MY_MODE) error {
	if conn.IsClosed() {
		return &SqlError{Num: 2006, Message: "Connection is closed"}
	}

	if C.my_query(&conn.c, &res.c, (*C.char)(stringPointer(sql)), C.ulong(len(sql)), mode) != 0 {
		return conn.lastError(sql)
	}

	return nil
}

func (conn *Connection) query(sql string, res *connQueryResult, mode C.MY_MODE) error {
	err := conn.execute(sql, &res.connResult, mode)
	if err != nil {
		return err
	}
	res.conn = conn
	res.fillFields()
	return nil
}
