package oursql

/*
#include "oursql.h"
*/
import "C"
import "unsafe"

var (
	c_TRUE  = C.my_bool(1)
	c_FALSE = C.my_bool(0)
)

type Stmt struct {
	s        C.OUR_STMT
	sql      string
	binds    []C.MYSQL_BIND
	bind_pos int
}

func (conn *Connection) Prepare(sql string) (*Stmt, error) {
	stmt := &Stmt{}
	stmt.sql = sql

	if C.our_prepare(&stmt.s, &conn.c, (*C.char)(stringPointer(sql)), C.ulong(len(sql))) != 0 {
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

func (stmt *Stmt) Execute() error {
	if C.our_stmt_execute(&stmt.s, &stmt.binds[0]) != 0 {
		return stmt.lastError()
	}
	return nil
}

func (stmt *Stmt) Close() {
	C.our_stmt_close(&stmt.s)
}

func cbool(gobool bool) *C.my_bool {
	if gobool {
		return &c_TRUE
	}
	return &c_FALSE
}
