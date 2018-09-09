package plan

import (
	"time"

	"github.com/modood/aip/db"
	"github.com/modood/aip/huobi"
	"github.com/modood/aip/util"

	"github.com/pkg/errors"
)

// Plan 定投计划接口定义
type Plan interface {
	Period() Period // 获取定投周期
	Invest() error  // 执行一次投资
	Monitor() error // 执行一次监控
}

type state struct {
	position   float64 // 持仓总额（基础货币）
	investment float64 // 投入总额（报价货币）
	price      float64 // 当前价格
	equity     float64 // 净值总额（报价货币）
	updated    uint64  // 更新时间
}

type plan struct {
	state
	client *huobi.Client // 火币客户端
	period Period        // 定投周期
	symbol string        // 交易品种
	amount float64       // 每期金额
}

// addOrder 新增订单
func (p *plan) addOrder(order *huobi.OpenOrder) error {
	return db.AddOrder(&db.Order{
		ID:          order.ID,
		Symbol:      order.Symbol,
		Type:        order.Type,
		Price:       order.FieldCashAmount / order.FieldAmount,
		BaseAmount:  order.FieldAmount,
		QuoteAmount: order.FieldCashAmount,
		Created:     order.CreatedAt / 1000,
	})
}

// addStatistics 新增统计
func (p *plan) addStatistics() error {
	return db.AddStatistics(&db.Statistics{
		Symbol:     p.symbol,
		Position:   p.state.position,
		Investment: p.state.investment,
		Price:      p.state.price,
		Equity:     p.state.equity,
		Created:    p.state.updated,
	})
}

// stateFlush 刷新，查询最新报价刷新净值数据
func (p *plan) stateFlush() error {
	price, err := p.client.SymbolPrice(p.symbol)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	p.state.price = price
	p.state.equity = price * p.state.position
	p.state.updated = uint64(time.Now().Unix())

	return nil
}

// stateInit 初始化，将统计数据从数据库加载到内存中
func (p *plan) stateInit() error {
	position, investment, err := db.OrderSummary()
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	p.state.position = position
	p.state.investment = investment
	p.state.updated = uint64(time.Now().Unix())

	if err = p.stateFlush(); err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	return nil
}

// stateUpdate 更新，根据订单更新状态
func (p *plan) stateUpdate(order *huobi.OpenOrder) error {
	p.state.position += order.FieldAmount
	p.state.investment += order.FieldCashAmount
	p.state.updated = uint64(time.Now().Unix())

	if err := p.stateFlush(); err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	return nil
}

// New 新建一个定投计划
func New(symbol string, amount float64,
	period Period, client *huobi.Client) (Plan, error) {

	p := &plan{
		client: client,
		period: period,
		symbol: symbol,
		amount: amount,
	}

	if err := p.stateInit(); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	return p, nil
}

// Period 获取定投周期
func (p *plan) Period() Period {
	return p.period
}

// Invest 执行一次投资
func (p *plan) Invest() error {
	order, err := p.client.Trade(p.symbol, huobi.BuyLimit, p.amount, -1)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	switch huobi.TradeType(order.Type) {
	case huobi.SellMarket, huobi.SellLimit:
		order.FieldAmount = -order.FieldAmount
		order.FieldCashAmount = -order.FieldCashAmount
	}

	if err = p.addOrder(order); err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	if err = p.stateUpdate(order); err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	return nil
}

// Monitor 执行一次监控
func (p *plan) Monitor() error {
	var err error

	if err = p.stateFlush(); err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	if err = p.addStatistics(); err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	// TODO 赎回检查

	return nil
}
