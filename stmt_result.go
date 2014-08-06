package mysql

/*
#include "cgo.h"
*/
import "C"

type stmtResult struct {
	c C.MY_STMT_RES
}

func (res *stmtResult) RowsAffected() uint64 {
	return uint64(res.c.affected_rows)
}

func (res *stmtResult) InsertId() uint64 {
	return uint64(res.c.insert_id)
}

type stmtQueryResult struct {
	stmtResult
	stmt   *Stmt
	fields []Field
}

func (res *stmtQueryResult) fillFields() {
	res.fields = fetchFields(res.stmt.s.meta)
}

func (res *stmtQueryResult) fetchNext() (row []Value, err error) {
	crow := C.my_stmt_fetch_next(&res.c)
	if crow.has_error != 0 {
		return nil, res.stmt.lastError()
	}

	return fetchNext(res.stmt.s.meta, crow, true)
}

func (res *stmtQueryResult) close() {
	C.my_stmt_close_result(&res.c)
}

func (res *stmtQueryResult) Fields() []Field {
	return res.fields
}

func (res *stmtQueryResult) IndexOf(name string) int {
	for i, field := range res.fields {
		if field.Name == name {
			return i
		}
	}
	return -1
}

type stmtDataTable struct {
	stmtQueryResult
	rows [][]Value
}

func (res *stmtDataTable) fillRows(stmt *Stmt) (err error) {
	rowCount := int(res.c.affected_rows)
	if rowCount == 0 {
		return nil
	}

	if rowCount < 0 {
		return stmt.lastError()
	}

	rows := make([][]Value, rowCount)
	for i := 0; i < rowCount; i++ {
		rows[i], err = res.fetchNext()
		if err != nil {
			return err
		}
	}

	res.rows = rows

	return nil
}

func (res *stmtDataTable) Rows() [][]Value {
	return res.rows
}

type stmtDataReader struct {
	stmtQueryResult
}

func (res *stmtDataReader) FetchNext() ([]Value, error) {
	return res.fetchNext()
}

func (res *stmtDataReader) Close() {
	res.close()
}
