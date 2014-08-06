package mysql

/*
#include "cgo.h"
*/
import "C"
import "fmt"

type SqlError struct {
	Num     int
	Message string
	Query   string
}

func (se *SqlError) Error() string {
	if se.Query == "" {
		return fmt.Sprintf("%v (errno %v)", se.Message, se.Num)
	}
	return fmt.Sprintf("%v (errno %v) during query: %s", se.Message, se.Num, se.Query)
}

func (se *SqlError) Number() int {
	return se.Num
}

func (conn *Connection) lastError(query string) error {
	if err := C.my_error(&conn.c); *err != 0 {
		return &SqlError{Num: int(C.my_errno(&conn.c)), Message: C.GoString(err), Query: query}
	}
	return &SqlError{0, "Unknow", string(query)}
}

type StmtError struct {
	Num     int
	Message string
	Stmt    *Stmt
}

func (self *StmtError) Error() string {
	if len(self.Stmt.sql) == 0 {
		return fmt.Sprintf("%v (errno %v)", self.Message, self.Num)
	}
	return fmt.Sprintf("%v (errno %v) during query: %s", self.Message, self.Num, self.Stmt.sql)
}

func (stmt *Stmt) lastError() error {
	if err := C.my_stmt_error(&stmt.s); *err != 0 {
		return &StmtError{Num: int(C.my_stmt_errno(&stmt.s)), Message: C.GoString(C.my_stmt_error(&stmt.s)), Stmt: stmt}
	}
	return &StmtError{0, "Unknow", stmt}
}
