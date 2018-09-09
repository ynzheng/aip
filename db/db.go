package db

import (
	"database/sql"

	"github.com/modood/aip/util"

	_ "github.com/mattn/go-sqlite3" // justifying
	"github.com/pkg/errors"
)

var db *sql.DB

// Order 订单表
type Order struct {
	ID          uint64  // 订单号
	Symbol      string  // 交易品种
	Type        string  // 交易类型
	Price       float64 // 成交价格
	BaseAmount  float64 // 成交金额（基础货币）
	QuoteAmount float64 // 花费金额（报价货币）
	Created     uint64  // 创建时间
}

const sqlOrder = `
CREATE TABLE IF NOT EXISTS 'orders' (
    'id'            INTEGER PRIMARY KEY,
    'symbol'        TEXT NOT NULL,
    'type'          TEXT NOT NULL,
    'price'         REAL NOT NULL,
    'base_amount'   REAL NOT NULL,
    'quote_amount'  REAL NOT NULL,
    'created'       TIMESTAMP default (datetime('now', 'localtime'))
);
`

// Statistics 统计表
type Statistics struct {
	ID         uint64  // 编号
	Symbol     string  // 交易品种
	Position   float64 // 持仓总额（基础货币）
	Investment float64 // 投入总额（报价货币）
	Price      float64 // 当前价格
	Equity     float64 // 净值总额（报价货币）
	Created    uint64  // 创建时间
}

const sqlStatistics = `
CREATE TABLE IF NOT EXISTS 'statistics' (
    'id'            INTEGER PRIMARY KEY,
    'symbol'        TEXT NOT NULL,
    'position'      REAL NOT NULL,
    'investment'    REAL NOT NULL,
    'price'         REAL NOT NULL,
    'equity'        REAL NOT NULL,
    'created'       TIMESTAMP default (datetime('now', 'localtime'))
);
`

// Init 初始化 sqlite3 数据库
func Init(path string) error {
	var err error

	db, err = sql.Open("sqlite3", path)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	if _, err = db.Exec(sqlOrder); err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	if _, err = db.Exec(sqlStatistics); err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	return nil
}

// AddOrder 新增订单
func AddOrder(order *Order) error {
	stmt, err := db.Prepare(`
		INSERT INTO
		orders(id, symbol, type, price, base_amount, quote_amount, created)
		VALUES(?, ?, ?, ?, ?, ?, datetime(?, 'unixepoch', 'localtime'));`)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	if _, err = stmt.Exec(
		order.ID,
		order.Symbol,
		order.Type,
		order.Price,
		order.BaseAmount,
		order.QuoteAmount,
		order.Created); err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	return nil
}

// AddStatistics 新增统计
func AddStatistics(statistics *Statistics) error {
	stmt, err := db.Prepare(`
		INSERT INTO
		statistics(symbol, position, investment, price, equity)
		VALUES(?,?,?,?,?);`)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	if _, err = stmt.Exec(
		statistics.Symbol,
		statistics.Position,
		statistics.Investment,
		statistics.Price,
		statistics.Equity); err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	return nil
}

// OrderSummary 返回订单汇总
// 返回值 position   持仓总额（基础货币）
// 返回值 investment 投入总额（报价货币）
func OrderSummary() (position, investment float64, err error) {
	row := db.QueryRow(`SELECT
		SUM(base_amount) AS position,
		SUM(quote_amount) AS investment FROM orders;`)
	if err := row.Scan(&position, &investment); err != nil {
		return 0, 0, errors.Wrap(err, util.FuncName())
	}

	return position, investment, nil
}
