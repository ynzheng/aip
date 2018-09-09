package plan

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDaily(t *testing.T) {
	Convey("should new daily successfully", t, func() {
		p, err := NewDaily(7, 56, 34)
		So(err, ShouldBeNil)

		r := p.Schedule()
		So(r, ShouldEqual, "34 56 7 * * *")
	})
}

func TestWeekly(t *testing.T) {
	Convey("should new weekly successfully", t, func() {
		p, err := NewWeekly(time.Saturday, 7, 56, 34)
		So(err, ShouldBeNil)

		r := p.Schedule()
		So(r, ShouldEqual, "34 56 7 * * 6")
	})
}

func TestMonthly(t *testing.T) {
	Convey("should new monthly successfully", t, func() {
		p, err := NewMonthly(1, 7, 56, 34)
		So(err, ShouldBeNil)

		r := p.Schedule()
		So(r, ShouldEqual, "34 56 7 1 * *")
	})
}
