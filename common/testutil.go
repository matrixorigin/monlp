package common

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"testing"
)

func ProjectPath(pathstrs ...string) string {
	_, fn, _, _ := runtime.Caller(0)
	fpath := path.Dir(fn)
	fpath = path.Join(fpath, "..")
	for _, pathstr := range pathstrs {
		fpath = path.Join(fpath, pathstr)
	}
	return fpath
}

func OpenFileForTest(pathstrs ...string) (*os.File, error) {
	rp := ProjectPath(pathstrs...)
	return os.Open(rp)
}

func CreateFileForTest(pathstrs ...string) (*os.File, error) {
	rp := ProjectPath(pathstrs...)
	return os.Create(rp)
}

func Assert(t *testing.T, condition bool, format string, args ...interface{}) {
	if !condition {
		if t != nil {
			t.Errorf(format, args...)
		} else {
			panic("Assert failed: " + fmt.Sprintf(format, args...))
		}
	}
}

func PanicAssert(t *testing.T, condition bool, format string, args ...interface{}) {
	if !condition {
		if t != nil {
			t.Fatalf(format, args...)
		} else {
			panic("Assert failed: " + fmt.Sprintf(format, args...))
		}
	}
}
