package mysql

import (
	"reflect"
	"unsafe"
)

// returns &s[0], which is not allowed in go
func stringPointer(s string) unsafe.Pointer {
	p := (*reflect.StringHeader)(unsafe.Pointer(&s))
	return unsafe.Pointer(p.Data)
}

// returns &b[0], which is not allowed in go
func bytePointer(b []byte) unsafe.Pointer {
	p := (*reflect.StringHeader)(unsafe.Pointer(&b))
	return unsafe.Pointer(p.Data)
}
