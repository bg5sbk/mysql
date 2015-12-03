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

// Prepared statement.
type Stmt struct {
	conn     *Connection
	s        C.MY_STMT
	sql      string
	binds    []C.MYSQL_BIND
	bind_pos int
}

// Prepare a statement.
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

// Clean bind parameters.
func (stmt *Stmt) CleanBind() {
	stmt.bind_pos = 0
}

// Number of input arguments.
func (stmt *Stmt) NumInput() int {
	return len(stmt.binds)
}

// Bind a int parameter.
func (stmt *Stmt) BindInt(value int32) {
	stmt.binds[stmt.bind_pos] = C.MYSQL_BIND{
		buffer_type: C.MYSQL_TYPE_LONG,
		buffer:      unsafe.Pointer(&value),
		is_null:     &c_FALSE,
	}
	stmt.bind_pos++
}

// Bind a tinyint parameter.
func (stmt *Stmt) BindTinyInt(value int8) {
	stmt.binds[stmt.bind_pos] = C.MYSQL_BIND{
		buffer_type: C.MYSQL_TYPE_TINY,
		buffer:      unsafe.Pointer(&value),
		is_null:     &c_FALSE,
	}
	stmt.bind_pos++
}

// Bind a smallint parameter.
func (stmt *Stmt) BindSmallInt(value int16) {
	stmt.binds[stmt.bind_pos] = C.MYSQL_BIND{
		buffer_type: C.MYSQL_TYPE_SHORT,
		buffer:      unsafe.Pointer(&value),
		is_null:     &c_FALSE,
	}
	stmt.bind_pos++
}

// Bind a bigint parameter.
func (stmt *Stmt) BindBigInt(value int64) {
	stmt.binds[stmt.bind_pos] = C.MYSQL_BIND{
		buffer_type: C.MYSQL_TYPE_LONGLONG,
		buffer:      unsafe.Pointer(&value),
		is_null:     &c_FALSE,
	}
	stmt.bind_pos++
}

// Bind a float parameter.
func (stmt *Stmt) BindFloat(value float32) {
	stmt.binds[stmt.bind_pos] = C.MYSQL_BIND{
		buffer_type: C.MYSQL_TYPE_FLOAT,
		buffer:      unsafe.Pointer(&value),
		is_null:     &c_FALSE,
	}
	stmt.bind_pos++
}

// Bind a double parameter.
func (stmt *Stmt) BindDouble(value float64) {
	stmt.binds[stmt.bind_pos] = C.MYSQL_BIND{
		buffer_type: C.MYSQL_TYPE_DOUBLE,
		buffer:      unsafe.Pointer(&value),
		is_null:     &c_FALSE,
	}
	stmt.bind_pos++
}

// Bind a text parameter.
func (stmt *Stmt) BindText(value string) {
	stmt.binds[stmt.bind_pos] = C.MYSQL_BIND{
		buffer_type:   C.MYSQL_TYPE_VAR_STRING,
		buffer:        stringPointer(value),
		buffer_length: (C.ulong)(len(value)),
		is_null:       &c_FALSE,
	}
	stmt.bind_pos++
}

// Bind a blob parameter.
func (stmt *Stmt) BindBlob(value []byte) {
	stmt.binds[stmt.bind_pos] = C.MYSQL_BIND{
		buffer_type:   C.MYSQL_TYPE_BLOB,
		buffer:        bytePointer(value),
		buffer_length: (C.ulong)(len(value)),
		is_null:       cbool(value == nil),
	}
	stmt.bind_pos++
}

func (stmt *Stmt) bind(paramType TypeCode, valuePtr unsafe.Pointer) {
	stmt.binds[stmt.bind_pos] = C.MYSQL_BIND{
		buffer_type: uint32(paramType),
		buffer:      valuePtr,
		is_null:     cbool(valuePtr == nil),
	}
	stmt.bind_pos++
}

// Bind parameter.
func (stmt *Stmt) Bind(value interface{}) {
	switch v := value.(type) {
	case int:
		stmt.BindBigInt(int64(v))
	case int8:
		stmt.BindTinyInt(v)
	case int16:
		stmt.BindSmallInt(v)
	case int32:
		stmt.BindInt(v)
	case int64:
		stmt.BindBigInt(v)
	case float32:
		stmt.BindFloat(v)
	case float64:
		stmt.BindDouble(v)
	case string:
		stmt.BindText(v)
	case []byte:
		stmt.BindBlob(v)
	case *int8:
		stmt.bind(C.MYSQL_TYPE_TINY, unsafe.Pointer(v))
	case *int16:
		stmt.bind(C.MYSQL_TYPE_SHORT, unsafe.Pointer(v))
	case *int32:
		stmt.bind(C.MYSQL_TYPE_LONG, unsafe.Pointer(v))
	case *int64:
		stmt.bind(C.MYSQL_TYPE_LONGLONG, unsafe.Pointer(v))
	case *float32:
		stmt.bind(C.MYSQL_TYPE_FLOAT, unsafe.Pointer(v))
	case *float64:
		stmt.bind(C.MYSQL_TYPE_DOUBLE, unsafe.Pointer(v))
	default:
		panic("unknow parameter type")
	}
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

// Execute statement as none-query.
func (stmt *Stmt) Execute() (Result, error) {
	res := &stmtResult{}

	if err := stmt.execute(res, C.MY_MODE_NONE); err != nil {
		return nil, err
	}

	return res, nil
}

// Execute statement and fill result into a DataTable.
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

// Execute statement and return a result reader. NOTE: Please remember close the reader.
func (stmt *Stmt) QueryReader() (DataReader, error) {
	res := &stmtDataReader{}

	if err := stmt.query(&res.stmtQueryResult, C.MY_MODE_READER); err != nil {
		return nil, err
	}

	return res, nil
}

// Close and dispose the statement.
func (stmt *Stmt) Close() error {
	if C.my_stmt_close(&stmt.s) != 0 {
		return stmt.lastError()
	}
	return nil
}

func cbool(gobool bool) *C.my_bool {
	if gobool {
		return &c_TRUE
	}
	return &c_FALSE
}
