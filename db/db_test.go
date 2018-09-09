package db

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInit(t *testing.T) {
	Convey("should init database successfully", t, func() {
		err := Init("/tmp/aip.sqlite3")
		So(err, ShouldBeNil)
	})
}

func TestAddOrder(t *testing.T) {
	Convey("should add order successfully", t, func() {
		err := AddOrder(&Order{
			ID:          uint64(time.Now().Unix()),
			Symbol:      "btcusdt",
			Type:        "buy-market",
			Price:       6432.463,
			BaseAmount:  0.321134,
			QuoteAmount: 2065.6825730419996,
			Created:     1536376845,
		})
		So(err, ShouldBeNil)
	})
}

func TestAddStatistics(t *testing.T) {
	Convey("should add statistics successfully", t, func() {
		err := AddStatistics(&Statistics{
			Symbol:     "btcusdt",
			Position:   1.34222223,
			Investment: 7633.79483225249,
			Price:      6432.463,
			Equity:     8633.79483225249,
		})
		So(err, ShouldBeNil)
	})
}

func TestOrderSummary(t *testing.T) {
	Convey("should return order summary successfully", t, func() {
		position, investment, err := OrderSummary()
		So(err, ShouldBeNil)
		So(position, ShouldNotEqual, 0)
		So(investment, ShouldNotEqual, 0)
	})
}
