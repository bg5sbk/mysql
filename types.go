package mysql

/*
#include "cgo.h"
*/
import "C"
import (
	"strconv"
)

var NULL = Value{}

type TypeCode int64

const (
	MYSQL_TYPE_TINY        = TypeCode(C.MYSQL_TYPE_TINY)        // TINYINT field
	MYSQL_TYPE_SHORT       = TypeCode(C.MYSQL_TYPE_SHORT)       // SMALLINT field
	MYSQL_TYPE_LONG        = TypeCode(C.MYSQL_TYPE_LONG)        // INTEGER field
	MYSQL_TYPE_INT24       = TypeCode(C.MYSQL_TYPE_INT24)       // MEDIUMINT field
	MYSQL_TYPE_LONGLONG    = TypeCode(C.MYSQL_TYPE_LONGLONG)    // BIGINT field
	MYSQL_TYPE_DECIMAL     = TypeCode(C.MYSQL_TYPE_DECIMAL)     // DECIMAL or NUMERIC field
	MYSQL_TYPE_NEWDECIMAL  = TypeCode(C.MYSQL_TYPE_NEWDECIMAL)  // Precision math DECIMAL or NUMERIC
	MYSQL_TYPE_FLOAT       = TypeCode(C.MYSQL_TYPE_FLOAT)       // FLOAT field
	MYSQL_TYPE_DOUBLE      = TypeCode(C.MYSQL_TYPE_DOUBLE)      // DOUBLE or REAL field
	MYSQL_TYPE_BIT         = TypeCode(C.MYSQL_TYPE_BIT)         // BIT field
	MYSQL_TYPE_TIMESTAMP   = TypeCode(C.MYSQL_TYPE_TIMESTAMP)   // TIMESTAMP field
	MYSQL_TYPE_DATE        = TypeCode(C.MYSQL_TYPE_DATE)        // DATE field
	MYSQL_TYPE_TIME        = TypeCode(C.MYSQL_TYPE_TIME)        // TIME field
	MYSQL_TYPE_DATETIME    = TypeCode(C.MYSQL_TYPE_DATETIME)    // DATETIME field
	MYSQL_TYPE_YEAR        = TypeCode(C.MYSQL_TYPE_YEAR)        // YEAR field
	MYSQL_TYPE_STRING      = TypeCode(C.MYSQL_TYPE_STRING)      // CHAR or BINARY field
	MYSQL_TYPE_VAR_STRING  = TypeCode(C.MYSQL_TYPE_VAR_STRING)  // VARCHAR or VARBINARY field
	MYSQL_TYPE_BLOB        = TypeCode(C.MYSQL_TYPE_BLOB)        // BLOB or TEXT field (use max_length to determine the maximum length)
	MYSQL_TYPE_SET         = TypeCode(C.MYSQL_TYPE_SET)         // SET field
	MYSQL_TYPE_ENUM        = TypeCode(C.MYSQL_TYPE_ENUM)        // ENUM field
	MYSQL_TYPE_GEOMETRY    = TypeCode(C.MYSQL_TYPE_GEOMETRY)    // Spatial field
	MYSQL_TYPE_NULL        = TypeCode(C.MYSQL_TYPE_NULL)        // NULL-type field
	MYSQL_TYPE_NEWDATE     = TypeCode(C.MYSQL_TYPE_NEWDATE)     //
	MYSQL_TYPE_TINY_BLOB   = TypeCode(C.MYSQL_TYPE_TINY_BLOB)   //
	MYSQL_TYPE_MEDIUM_BLOB = TypeCode(C.MYSQL_TYPE_MEDIUM_BLOB) //
	MYSQL_TYPE_LONG_BLOB   = TypeCode(C.MYSQL_TYPE_LONG_BLOB)
)

// Non-query result.
type Result interface {
	// Get how many rows affected by query.
	RowsAffected() int64

	// Get the last insert id of the query.
	InsertId() int64
}

// Query result.
type QueryResult interface {
	Result

	// Get result fields.
	Fields() []Field

	// Get field index.
	IndexOf(string) int
}

// Filled result.
type DataTable interface {
	QueryResult

	// Get all rows.
	Rows() [][]Value
}

// Result reader.
type DataReader interface {
	QueryResult

	// Fetch next row.
	FetchNext() ([]Value, error)

	// Close and dispose result.
	Close()
}

// Field described a column returned by mysql
type Field struct {
	Name string
	Type TypeCode
}

// Value can store any SQL value. NULL is stored as nil.
type Value struct {
	isStmtValue bool
	Type        TypeCode
	Inner       []byte
}

// Check value is null or not.
func (v *Value) IsNull() bool {
	return v.Inner == nil
}

// Auto convert.
func (v *Value) Interface() interface{} {
	if v.isStmtValue {
		switch v.Type {
		case MYSQL_TYPE_TINY, MYSQL_TYPE_YEAR, MYSQL_TYPE_SHORT, MYSQL_TYPE_INT24, MYSQL_TYPE_LONG, MYSQL_TYPE_LONGLONG:
			return v.getInt()
		case MYSQL_TYPE_FLOAT, MYSQL_TYPE_DOUBLE:
			return v.getFloat()
		}
	}
	return string(v.Inner)
}

// Convert to int8 value.
func (v *Value) Int8() int8 {
	return int8(v.getInt())
}

// Convert to int16 value.
func (v *Value) Int16() int16 {
	return int16(v.getInt())
}

// Convert to int32 value.
func (v *Value) Int32() int32 {
	return int32(v.getInt())
}

// Convert to int64 value.
func (v *Value) Int64() int64 {
	return v.getInt()
}

// Convert to float32 value.
func (v *Value) Float32() float32 {
	return float32(v.getFloat())
}

// Convert to float64 value.
func (v *Value) Float64() float64 {
	return v.getFloat()
}

// Convert to int value.
func (v *Value) getInt() int64 {
	if v.isStmtValue {
		switch v.Type {
		case MYSQL_TYPE_TINY:
			return int64(*(*int8)(bytePointer(v.Inner)))
		case MYSQL_TYPE_YEAR:
			fallthrough
		case MYSQL_TYPE_SHORT:
			return int64(*(*int16)(bytePointer(v.Inner)))
		case MYSQL_TYPE_INT24:
			fallthrough
		case MYSQL_TYPE_LONG:
			return int64(*(*int32)(bytePointer(v.Inner)))
		case MYSQL_TYPE_LONGLONG:
			return *(*int64)(bytePointer(v.Inner))
		default:
			panic("the value is not integer type")
		}
	}
	r, err := strconv.ParseInt(byteString(v.Inner), 0, 64)
	if err != nil {
		panic(err)
	}
	return r
}

// Convert to float value.
func (v *Value) getFloat() float64 {
	if v.isStmtValue {
		switch v.Type {
		case MYSQL_TYPE_FLOAT:
			return float64(*(*float32)(bytePointer(v.Inner)))
		case MYSQL_TYPE_DOUBLE:
			return *(*float64)(bytePointer(v.Inner))
		default:
			panic("the value is not float type")
		}
	}
	r, err := strconv.ParseFloat(byteString(v.Inner), 64)
	if err != nil {
		panic(err)
	}
	return r
}

// Convert to string value.
func (v *Value) String() string {
	if v.isStmtValue {
		switch v.Type {
		// string
		case MYSQL_TYPE_DECIMAL:
			fallthrough
		case MYSQL_TYPE_NEWDATE:
			fallthrough
		case MYSQL_TYPE_NEWDECIMAL:
			fallthrough
		case MYSQL_TYPE_STRING:
			fallthrough
		case MYSQL_TYPE_VAR_STRING:
			fallthrough
		case MYSQL_TYPE_TINY_BLOB:
			fallthrough
		case MYSQL_TYPE_BLOB:
			fallthrough
		case MYSQL_TYPE_MEDIUM_BLOB:
			fallthrough
		case MYSQL_TYPE_LONG_BLOB:
			fallthrough
		case MYSQL_TYPE_BIT:
			return string(v.Inner)
		// int
		case MYSQL_TYPE_TINY:
			fallthrough
		case MYSQL_TYPE_YEAR:
			fallthrough
		case MYSQL_TYPE_SHORT:
			fallthrough
		case MYSQL_TYPE_INT24:
			fallthrough
		case MYSQL_TYPE_LONG:
			fallthrough
		case MYSQL_TYPE_LONGLONG:
			return strconv.FormatInt(v.getInt(), 10)
		// float
		case MYSQL_TYPE_FLOAT:
			fallthrough
		case MYSQL_TYPE_DOUBLE:
			return strconv.FormatFloat(v.getFloat(), 'f', -1, 64)
		}
	}
	return string(v.Inner)
}
