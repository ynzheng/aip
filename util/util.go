package util

import (
	"path"
	"runtime"
)

// FuncName get current scope of function name
func FuncName() string {
	p := make([]uintptr, 1)
	runtime.Callers(2, p)
	fullname := runtime.FuncForPC(p[0]).Name()

	_, name := path.Split(fullname)
	return name
}
