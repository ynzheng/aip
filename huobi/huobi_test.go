package huobi

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewClient(t *testing.T) {
	Convey("should new client successfully", t, func() {
		c, err := NewClient("https://api.huobi.pro", "apikey", "apisecret")
		So(err, ShouldBeNil)
		So(c.key, ShouldEqual, "apikey")
		So(c.secret, ShouldEqual, "apisecret")
	})
}

func TestAccounts(t *testing.T) {
	Convey("should return all accounts successfully", t, func() {
		c, err := NewClient("https://api.huobi.pro", "apikey", "apisecret")
		So(err, ShouldBeNil)

		r, err := c.Accounts()
		So(err, ShouldBeNil)
		So(r, ShouldNotBeEmpty)
	})
}

func TestSpotAccount(t *testing.T) {
	Convey("should return spot account successfully", t, func() {
		c, err := NewClient("https://api.huobi.pro", "apikey", "apisecret")
		So(err, ShouldBeNil)

		r, err := c.SpotAccount()
		So(err, ShouldBeNil)
		So(r.List, ShouldNotBeEmpty)
	})
}

func TestSpotAccountID(t *testing.T) {
	Convey("should return spot account id successfully", t, func() {
		c, err := NewClient("https://api.huobi.pro", "apikey", "apisecret")
		So(err, ShouldBeNil)

		r, err := c.SpotAccountID()
		So(err, ShouldBeNil)
		So(r, ShouldNotEqual, 0)
	})
}

func TestSpotAccountBalance(t *testing.T) {
	Convey("should return spot account balance successfully", t, func() {
		c, err := NewClient("https://api.huobi.pro", "apikey", "apisecret")
		So(err, ShouldBeNil)

		r, err := c.SpotAccountBalance("btc")
		So(err, ShouldBeNil)
		So(r, ShouldNotEqual, 0)
	})
}

func TestSymbol(t *testing.T) {
	Convey("should return symbol successfully", t, func() {
		c, err := NewClient("https://api.huobi.pro", "apikey", "apisecret")
		So(err, ShouldBeNil)

		r, err := c.Symbol("btcusdt")
		So(err, ShouldBeNil)
		So(r, ShouldNotBeNil)
	})
}

func TestSymbolPrice(t *testing.T) {
	Convey("should return symbol price successfully", t, func() {
		c, err := NewClient("https://api.huobi.pro", "apikey", "apisecret")
		So(err, ShouldBeNil)

		r, err := c.SymbolPrice("btcusdt")
		So(err, ShouldBeNil)
		So(r, ShouldBeGreaterThan, 0)
	})
}

func TestSymbols(t *testing.T) {
	Convey("should return all support symbol successfully", t, func() {
		c, err := NewClient("https://api.huobi.pro", "apikey", "apisecret")
		So(err, ShouldBeNil)

		r, err := c.Symbols()
		So(err, ShouldBeNil)
		So(r, ShouldNotBeEmpty)
	})
}

func TestOpenOrder(t *testing.T) {
	Convey("should return all support symbol successfully", t, func() {
		c, err := NewClient("https://api.huobi.pro", "apikey", "apisecret")
		So(err, ShouldBeNil)

		r, err := c.OpenOrder(2542019603)
		So(err, ShouldBeNil)
		So(r, ShouldNotBeEmpty)
	})
}

func TestTrade(t *testing.T) {
	Convey("should trade successfully", t, func() {
		c, err := NewClient("https://api.huobi.pro", "apikey", "apisecret")
		So(err, ShouldBeNil)

		r, err := c.Trade("btcusdt", BuyLimit, 10, 1)
		So(err, ShouldBeNil)
		So(r, ShouldNotBeEmpty)
	})
}
