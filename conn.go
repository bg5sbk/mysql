package mysql

/*
#include "cgo.h"
*/
import "C"
import (
	"unsafe"
)

func init() {
	// This needs to be called before threads begin to spawn.
	C.my_library_init()
}

type ConnectionParams struct {
	Host       string
	Port       int
	Uname      string
	Pass       string
	DbName     string
	UnixSocket string
	Charset    string
	Flags      uint64
}

func (c *ConnectionParams) EnableMultiStatements() {
	c.Flags |= C.CLIENT_MULTI_STATEMENTS
}

func (c *ConnectionParams) Redact() {
	c.Pass = "***"
}

type Connection struct {
	c      C.MYSQL
	closed bool
}

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

func (conn *Connection) Id() int64 {
	return int64(C.my_thread_id(&conn.c))
}

func (conn *Connection) Close() {
	C.my_close(&conn.c)
	conn.closed = true
}

func (conn *Connection) IsClosed() bool {
	return conn.closed
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

func (conn *Connection) Execute(sql string) (Result, error) {
	res := &connResult{}

	err := conn.execute(sql, res, C.MY_MODE_NONE)
	if err != nil {
		return nil, err
	}

	return res, nil
}

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
