package mysql

/*
#include "cgo.h"
*/
import "C"
import "unsafe"

var (
	c_TRUE  = C.my_bool(1)
	c_FALSE = C.my_bool(0)
)

type Stmt struct {
	conn     *Connection
	s        C.MY_STMT
	sql      string
	binds    []C.MYSQL_BIND
	bind_pos int
}

func (conn *Connection) Prepare(sql string) (*Stmt, error) {
	stmt := &Stmt{}
	stmt.conn = conn
	stmt.sql = sql

	if C.my_prepare(&stmt.s, &conn.c, (*C.char)(stringPointer(sql)), C.ulong(len(sql))) != 0 {
		return nil, conn.lastError(sql)
	}

	stmt.binds = make([]C.MYSQL_BIND, int(stmt.s.param_count))

	return stmt, nil
}

func (stmt *Stmt) CleanBind() {
	stmt.bind_pos = 0
}

func (stmt *Stmt) BindInt(value int32) {
	stmt.binds[stmt.bind_pos].buffer_type = C.MYSQL_TYPE_LONG
	stmt.binds[stmt.bind_pos].buffer = unsafe.Pointer(&value)
	stmt.binds[stmt.bind_pos].is_null = &c_FALSE
	stmt.bind_pos++
}

func (stmt *Stmt) BindTinyInt(value int8) {
	stmt.binds[stmt.bind_pos].buffer_type = C.MYSQL_TYPE_TINY
	stmt.binds[stmt.bind_pos].buffer = unsafe.Pointer(&value)
	stmt.binds[stmt.bind_pos].is_null = &c_FALSE
	stmt.bind_pos++
}

func (stmt *Stmt) BindSmallInt(value int16) {
	stmt.binds[stmt.bind_pos].buffer_type = C.MYSQL_TYPE_SHORT
	stmt.binds[stmt.bind_pos].buffer = unsafe.Pointer(&value)
	stmt.binds[stmt.bind_pos].is_null = &c_FALSE
	stmt.bind_pos++
}

func (stmt *Stmt) BindBigInt(value int64) {
	stmt.binds[stmt.bind_pos].buffer_type = C.MYSQL_TYPE_LONGLONG
	stmt.binds[stmt.bind_pos].buffer = unsafe.Pointer(&value)
	stmt.binds[stmt.bind_pos].is_null = &c_FALSE
	stmt.bind_pos++
}

func (stmt *Stmt) BindFloat(value float32) {
	stmt.binds[stmt.bind_pos].buffer_type = C.MYSQL_TYPE_FLOAT
	stmt.binds[stmt.bind_pos].buffer = unsafe.Pointer(&value)
	stmt.binds[stmt.bind_pos].is_null = &c_FALSE
	stmt.bind_pos++
}

func (stmt *Stmt) BindDouble(value float64) {
	stmt.binds[stmt.bind_pos].buffer_type = C.MYSQL_TYPE_DOUBLE
	stmt.binds[stmt.bind_pos].buffer = unsafe.Pointer(&value)
	stmt.binds[stmt.bind_pos].is_null = &c_FALSE
	stmt.bind_pos++
}

func (stmt *Stmt) BindString(value string) {
	stmt.binds[stmt.bind_pos].buffer_type = C.MYSQL_TYPE_VAR_STRING
	stmt.binds[stmt.bind_pos].buffer = stringPointer(value)
	stmt.binds[stmt.bind_pos].buffer_length = (C.ulong)(len(value))
	stmt.binds[stmt.bind_pos].is_null = &c_FALSE
	stmt.bind_pos++
}

func (stmt *Stmt) BindBlob(value []byte) {
	stmt.binds[stmt.bind_pos].buffer_type = C.MYSQL_TYPE_BLOB
	stmt.binds[stmt.bind_pos].buffer = bytePointer(value)
	stmt.binds[stmt.bind_pos].buffer_length = (C.ulong)(len(value))
	stmt.binds[stmt.bind_pos].is_null = cbool(value == nil)
	stmt.bind_pos++
}

func (stmt *Stmt) Bind(paramType TypeCode, valuePtr unsafe.Pointer, length int) {
	stmt.binds[stmt.bind_pos].buffer_type = uint32(paramType)
	stmt.binds[stmt.bind_pos].buffer = valuePtr
	stmt.binds[stmt.bind_pos].buffer_length = (C.ulong)(length)
	stmt.binds[stmt.bind_pos].is_null = cbool(valuePtr == nil)
	stmt.bind_pos++
}

func (stmt *Stmt) execute(res *stmtResult, mode C.MY_MODE) error {
	if stmt.conn.IsClosed() {
		return &SqlError{Num: 2006, Message: "Connection is closed"}
	}

	var bind *C.MYSQL_BIND
	if len(stmt.binds) > 0 {
		bind = &stmt.binds[0]
	}

	if C.my_stmt_execute(&stmt.s, bind, &res.c, mode) != 0 {
		return stmt.lastError()
	}

	return nil
}

func (stmt *Stmt) query(res *stmtQueryResult, mode C.MY_MODE) error {
	if err := stmt.execute(&res.stmtResult, mode); err != nil {
		return err
	}
	res.stmt = stmt
	res.fillFields()
	return nil
}

func (stmt *Stmt) Execute() (Result, error) {
	res := &stmtResult{}

	if err := stmt.execute(res, C.MY_MODE_NONE); err != nil {
		return nil, err
	}

	return res, nil
}

func (stmt *Stmt) QueryTable() (DataTable, error) {
	res := &stmtDataTable{}

	if err := stmt.query(&res.stmtQueryResult, C.MY_MODE_TABLE); err != nil {
		return nil, err
	}
	defer res.close()

	if err := res.fillRows(stmt); err != nil {
		return nil, err
	}

	return res, nil
}

func (stmt *Stmt) QueryReader() (DataReader, error) {
	res := &stmtDataReader{}

	if err := stmt.query(&res.stmtQueryResult, C.MY_MODE_READER); err != nil {
		return nil, err
	}

	return res, nil
}

func (stmt *Stmt) Close() {
	C.my_stmt_close(&stmt.s)
}

func cbool(gobool bool) *C.my_bool {
	if gobool {
		return &c_TRUE
	}
	return &c_FALSE
}
