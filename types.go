package oursql

/*
#include "oursql.h"
*/
import "C"
import "strconv"

var NULL = Value{}

type TypeCode int64

const (
	MYSQL_TYPE_TINY       = TypeCode(C.MYSQL_TYPE_TINY)       // TINYINT field
	MYSQL_TYPE_SHORT      = TypeCode(C.MYSQL_TYPE_SHORT)      // SMALLINT field
	MYSQL_TYPE_LONG       = TypeCode(C.MYSQL_TYPE_LONG)       // INTEGER field
	MYSQL_TYPE_INT24      = TypeCode(C.MYSQL_TYPE_INT24)      // MEDIUMINT field
	MYSQL_TYPE_LONGLONG   = TypeCode(C.MYSQL_TYPE_LONGLONG)   // BIGINT field
	MYSQL_TYPE_DECIMAL    = TypeCode(C.MYSQL_TYPE_DECIMAL)    // DECIMAL or NUMERIC field
	MYSQL_TYPE_NEWDECIMAL = TypeCode(C.MYSQL_TYPE_NEWDECIMAL) // Precision math DECIMAL or NUMERIC
	MYSQL_TYPE_FLOAT      = TypeCode(C.MYSQL_TYPE_FLOAT)      // FLOAT field
	MYSQL_TYPE_DOUBLE     = TypeCode(C.MYSQL_TYPE_DOUBLE)     // DOUBLE or REAL field
	MYSQL_TYPE_BIT        = TypeCode(C.MYSQL_TYPE_BIT)        // BIT field
	MYSQL_TYPE_TIMESTAMP  = TypeCode(C.MYSQL_TYPE_TIMESTAMP)  // TIMESTAMP field
	MYSQL_TYPE_DATE       = TypeCode(C.MYSQL_TYPE_DATE)       // DATE field
	MYSQL_TYPE_TIME       = TypeCode(C.MYSQL_TYPE_TIME)       // TIME field
	MYSQL_TYPE_DATETIME   = TypeCode(C.MYSQL_TYPE_DATETIME)   // DATETIME field
	MYSQL_TYPE_YEAR       = TypeCode(C.MYSQL_TYPE_YEAR)       // YEAR field
	MYSQL_TYPE_STRING     = TypeCode(C.MYSQL_TYPE_STRING)     // CHAR or BINARY field
	MYSQL_TYPE_VAR_STRING = TypeCode(C.MYSQL_TYPE_VAR_STRING) // VARCHAR or VARBINARY field
	MYSQL_TYPE_BLOB       = TypeCode(C.MYSQL_TYPE_BLOB)       // BLOB or TEXT field (use max_length to determine the maximum length)
	MYSQL_TYPE_SET        = TypeCode(C.MYSQL_TYPE_SET)        // SET field
	MYSQL_TYPE_ENUM       = TypeCode(C.MYSQL_TYPE_ENUM)       // ENUM field
	MYSQL_TYPE_GEOMETRY   = TypeCode(C.MYSQL_TYPE_GEOMETRY)   // Spatial field
	MYSQL_TYPE_NULL       = TypeCode(C.MYSQL_TYPE_NULL)       // NULL-type field
)

// Field described a column returned by mysql
type Field struct {
	Name string
	Type TypeCode
}

// Value can store any SQL value. NULL is stored as nil.
type Value struct {
	Type  TypeCode
	Inner []byte
}

func (v *Value) Int() int64 {
	r, err := strconv.ParseInt(string(v.Inner), 0, 64)
	if err != nil {
		panic(err)
	}
	return r
}

func (v *Value) Float() float64 {
	r, err := strconv.ParseFloat(string(v.Inner), 64)
	if err != nil {
		panic(err)
	}
	return r
}

func (v *Value) String() string {
	return string(v.Inner)
}

type Result interface {
	RowsAffected() uint64
	InsertId() uint64
}

type DataTable interface {
	Fields() []Field
	Rows() [][]Value
	IndexOf(string) int
}

type DataReader interface {
	Fields() []Field
	FetchNext() ([]Value, error)
	Close()
}
