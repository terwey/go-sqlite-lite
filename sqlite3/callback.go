package sqlite3

/*
#include "sqlite3.h"
#include "callback.h"

*/
import "C"
import (
	"unsafe"

	gopointer "github.com/mattn/go-pointer"
)

//export goExecCallback
func goExecCallback(callbackFn unsafe.Pointer, columnCount C.int, columnData, columnNames **C.char) C.int {
	var fn ExecCallback = gopointer.Restore(callbackFn).(ExecCallback)

	data := make([]string, 0, columnCount)
	for _, v := range (*[1 << 28]*C.char)(unsafe.Pointer(columnData))[:columnCount:columnCount] {
		data = append(data, goStr(v))
	}

	cols := make([]string, 0, columnCount)
	for _, v := range (*[1 << 28]*C.char)(unsafe.Pointer(columnNames))[:columnCount:columnCount] {
		cols = append(cols, goStr(v))
	}

	return C.int(fn(data, cols))
}

//export goExecDataCallback
func goExecDataCallback(callbackFn unsafe.Pointer, columnCount C.int, columnData, columnNames **C.char) C.int {
	var fn ExecDataCallback = gopointer.Restore(callbackFn).(ExecDataCallback)

	data := make([]string, 0, columnCount)
	for _, v := range (*[1 << 28]*C.char)(unsafe.Pointer(columnData))[:columnCount:columnCount] {
		data = append(data, goStr(v))
	}

	return C.int(fn(data))
}

type ExecCallback func(columnData, columnNames []string) int
type ExecDataCallback func(columnData []string) int

// ExecWithCallback allows the caller to have a convenient method of iterating
// over results without having to Step through the rows himself. SQLite will keep calling the
// Callback function until no more data is available or a non-zero (SUCCESS) is returned.
//
// ExecWithCallback calls sqlite3_exec and binds a Callback function to it.
// The signature for the callback is: func(columnData, columnNames []string) int
//
// Implementation detail: The fourth parameter for sqlite3_exec will be a pointer to
// your provided function call. https://www.sqlite.org/c3ref/exec.html
//
// See ExecWithDataCallback if you don't need the slice of column names.
func (c *Conn) ExecWithCallback(sql string, callback ExecCallback) error {
	p := gopointer.Save(callback)
	defer gopointer.Unref(p)
	if rc := C.sqlite3_exec(c.db,
		cStr(sql+"\x00"),
		(*[0]byte)(C.cgoExecCallback),
		p,
		nil); rc != OK {
		return libErr(rc, c.db)
	}

	return nil
}

// ExecWithDataCallback allows the caller to have a convenient method of iterating
// over data only results without having to Step through the rows himself. SQLite
// will keep calling the Callback function until no more data is available or a
// non-zero (SUCCESS) is returned.
//
// ExecWithDataCallback calls sqlite3_exec and binds a Callback function to it.
// The signature for the callback is: func(columnData []string) int
//
// Implementation detail: The fourth parameter for sqlite3_exec will be a pointer to
// your provided function call. https://www.sqlite.org/c3ref/exec.html
//
// See ExecWithCallback if you also need the slice of column names.
func (c *Conn) ExecWithDataCallback(sql string, callback ExecDataCallback) error {
	p := gopointer.Save(callback)
	defer gopointer.Unref(p)
	if rc := C.sqlite3_exec(c.db,
		cStr(sql+"\x00"),
		(*[0]byte)(C.cgoExecDataCallback),
		p,
		nil); rc != OK {
		return libErr(rc, c.db)
	}

	return nil
}
