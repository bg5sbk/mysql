package oursql

/*
#include "oursql.h"
*/
import "C"

type stmtResult struct {
	c    C.OUR_STMT_RES
	stmt *Stmt
}

func (res *stmtResult) RowsAffected() uint64 {
	return uint64(res.c.affected_rows)
}

func (res *stmtResult) InsertId() uint64 {
	return uint64(res.c.insert_id)
}
