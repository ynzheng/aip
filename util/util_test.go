package util

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFuncName(t *testing.T) {
	name := FuncName()
	Convey("should return func name correctly", t, func(c C) {
		anonymity := FuncName()
		So(name, ShouldEqual, "util.TestFuncName")
		So(anonymity, ShouldEqual, "util.TestFuncName.func1")
	})
}
