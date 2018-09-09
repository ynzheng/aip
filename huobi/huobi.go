package huobi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/modood/aip/util"

	jsoniter "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

var (
	errSymbolPriceFailed = errors.New("get symbol price failed")
	errSymbolNotFound    = errors.New("symbol not found")
	errAccountNotFound   = errors.New("account not found")
	errUnkownTradeType   = errors.New("unknown trade type")

	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

// TradeType 交易类型
type TradeType string

// 交易类型
const (
	BuyMarket  TradeType = "buy-market"  // 市价买入
	SellMarket TradeType = "sell-market" // 市价卖出
	BuyLimit   TradeType = "buy-limit"   // 限价买入
	SellLimit  TradeType = "sell-limit"  // 限价卖出
)

// Client 火币 API 客户端
type Client struct {
	host     string
	key      string
	secret   string
	symbols  []*Symbol
	accounts []*Account
}

// Account 账户信息
type Account struct {
	ID      uint64
	Type    string
	SubType string `mapstructure:"subtype" json:"subtype"`
	State   string
	List    []struct {
		Currency string
		Type     string
		Balance  float64
	}
}

// Symbol 交易品种
type Symbol struct {
	Symbol          string
	BaseCurrency    string `mapstructure:"base-currency" json:"base-currency"`
	QuoteCurrency   string `mapstructure:"quote-currency" json:"quote-currency"`
	PricePrecision  int    `mapstructure:"price-precision" json:"price-precision"`
	AmountPrecision int    `mapstructure:"amount-precision" json:"amount-precision"`
	SymbolPartition string `mapstructure:"symbol-partition" json:"symbol-partition"`
}

// OpenOrder 订单（状态可能未完成）
type OpenOrder struct {
	ID              uint64
	AccountID       uint64 `mapstructure:"account-id" json:"account-id"`
	Source          string
	Type            string
	State           string
	Symbol          string
	Amount          float64
	Price           float64
	FieldAmount     float64 `mapstructure:"field-amount" json:"field-amount"`
	FieldCashAmount float64 `mapstructure:"field-cash-amount" json:"field-cash-amount"`
	FieldFees       float64 `mapstructure:"field-fees" json:"field-fees"`
	CreatedAt       uint64  `mapstructure:"created-at" json:"created-at"`
	FinishedAt      uint64  `mapstructure:"finished-at" json:"finished-at"`
	CanceledAt      uint64  `mapstructure:"canceled-at" json:"canceled-at"`
}

// huobiError 火币 API 错误码
type huobiError struct {
	Status string
	// Error code:
	// base-symbol-error                            交易对不存在
	// base-currency-error                          币种不存在
	// base-date-error                              错误的日期格式
	// account-transfer-balance-insufficient-error  余额不足无法冻结
	// bad-argument                                 无效参数
	// api-signature-not-valid                      API签名错误
	// gateway-internal-error                       系统繁忙，请稍后再试
	// security-require-assets-password             需要输入资金密码
	// audit-failed                                 下单失败
	// ad-ethereum-addresss                         请输入有效的以太坊地址
	// order-accountbalance-error                   账户余额不足
	// order-limitorder-price-error                 限价单下单价格超出限制
	// order-limitorder-amount-error                限价单下单数量超出限制
	// order-orderprice-precision-error             下单价格超出精度限制
	// order-orderamount-precision-error            下单数量超过精度限制
	// order-marketorder-amount-error               下单数量超出限制
	// order-queryorder-invalid                     查询不到此条订单
	// order-orderstate-error                       订单状态错误
	// order-datelimit-error                        查询超出时间限制
	// order-update-error                           订单更新出错
	// bad-request                                  错误请求
	// invalid-parameter                            参数错
	// invalid-command                              指令错
	Code    string `mapstructure:"err-code" json:"err-code"`
	Message string `mapstructure:"err-msg" json:"err-msg"`
}

// NewClient 创建火币客户端
func NewClient(host, key, secret string) (*Client, error) {
	c := &Client{
		host:   host,
		key:    key,
		secret: secret,
	}

	symbols, err := c.Symbols()
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}
	c.symbols = symbols

	accounts, err := c.Accounts()
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}
	c.accounts = accounts

	return c, nil
}

// Symbols 返回火币支持的所有交易品种
func (c *Client) Symbols() ([]*Symbol, error) {
	if len(c.symbols) != 0 {
		return c.symbols, nil
	}

	m, err := c.req("GET", "/v1/common/symbols", nil)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	r := struct{ Data []*Symbol }{}
	if err = decode(m, &r); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	return r.Data, nil
}

// Symbol 根据名称获取交易品种
func (c *Client) Symbol(name string) (*Symbol, error) {
	for _, s := range c.symbols {
		if s.Symbol == name {
			return s, nil
		}
	}

	return nil, errors.Wrap(errSymbolNotFound, util.FuncName())
}

// SymbolPrice 根据名称获取交易品种的最新价格
func (c *Client) SymbolPrice(name string) (float64, error) {
	m, err := c.req("GET", "/market/trade",
		map[string]string{"symbol": name})
	if err != nil {
		return 0, errors.Wrap(err, util.FuncName())
	}

	r := struct {
		Tick struct{ Data []struct{ Price float64 } }
	}{}
	if err = decode(m, &r); err != nil {
		return 0, errors.Wrap(err, util.FuncName())
	}
	if len(r.Tick.Data) == 0 {
		return 0, errors.Wrap(errSymbolPriceFailed, util.FuncName())
	}

	return r.Tick.Data[0].Price, nil
}

// Accounts 返回当前用户的账户列表（包括现货期货等）
func (c *Client) Accounts() ([]*Account, error) {
	if len(c.accounts) != 0 {
		return c.accounts, nil
	}

	m, err := c.req("GET", "/v1/account/accounts", nil)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	r := struct{ Data []*Account }{}
	if err = decode(m, &r); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	return r.Data, nil
}

// SpotAccountID 返回现货账户 ID
func (c *Client) SpotAccountID() (uint64, error) {
	var id uint64
	for _, a := range c.accounts {
		if a.Type == "spot" {
			id = a.ID
			break
		}
	}
	if id == 0 {
		return 0, errors.Wrap(errAccountNotFound, util.FuncName())
	}

	return id, nil
}

// SpotAccount 返回现货账户
func (c *Client) SpotAccount() (*Account, error) {
	id, err := c.SpotAccountID()
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	m, err := c.req("GET", "/v1/account/accounts/"+strconv.FormatUint(id, 10)+"/balance", nil)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	r := struct{ Data *Account }{}
	if err = decode(m, &r); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	return r.Data, nil
}

// SpotAccountBalance 返回现货账户下指定货币的余额
func (c *Client) SpotAccountBalance(currency string) (float64, error) {
	a, err := c.SpotAccount()
	if err != nil {
		return 0, errors.Wrap(err, util.FuncName())
	}

	for _, ccy := range a.List {
		if ccy.Currency == currency && ccy.Type == "trade" {
			return ccy.Balance, nil
		}
	}

	return 0, nil
}

// OpenOrder 根据 ID 查看订单信息
func (c *Client) OpenOrder(ID uint64) (*OpenOrder, error) {
	m, err := c.req("GET", "/v1/order/orders/"+strconv.FormatUint(ID, 10), nil)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	r := struct{ Data *OpenOrder }{}
	if err = decode(m, &r); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	return r.Data, nil
}

// Trade 发起一笔交易
// 参数 amount 限价单表示下单数量，市价买单时表示买多少钱，市价卖单时表示卖多少币
// 参数 price  限价单表示报价，市价单会忽略掉该参数
func (c *Client) Trade(symbol string, cmd TradeType, amount, price float64) (*OpenOrder, error) {
	s, err := c.Symbol(symbol)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	id, err := c.SpotAccountID()
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	params := map[string]string{
		"account-id": strconv.FormatUint(id, 10),
		"source":     "api",
		"symbol":     s.Symbol,
		"amount":     floor(amount, s.AmountPrecision),
		"type":       string(cmd),
	}

	if cmd == BuyLimit || cmd == SellLimit {
		params["price"] = floor(price, s.PricePrecision)
	}

	m, err := c.req("POST", "/v1/order/orders/place", params)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	r := struct{ Data uint64 }{}
	if err := decode(m, &r); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	time.Sleep(time.Second * 5) // await until order state changed: submitted => filled
	o, err := c.OpenOrder(r.Data)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	return o, nil
}

// querystring 格式化请求参数
func querystring(m map[string]string) string {
	l := len(m)

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	q := make([]string, l)
	for i, k := range keys {
		q[i] = url.QueryEscape(k) + "=" + url.QueryEscape(m[k])
	}

	return strings.Join(q, "&")
}

// floor 向下取指定精度的字符串数字
func floor(f float64, prec int) string {
	i := math.Pow10(prec)
	return strconv.FormatFloat(math.Floor(f*i)/i, 'f', -1, 64)
}

// handle 处理火币错误码
func handle(bs []byte, err error) error {
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(bs, &m)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	e := huobiError{}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &e,
	})
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	err = decoder.Decode(m)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	if e.Status != "ok" {
		err = fmt.Errorf("Code: %s, %s", e.Code, e.Message)
		return errors.Wrap(err, util.FuncName())
	}

	return nil
}

// sign 签名
func (c *Client) sign(content string) (string, error) {
	h := hmac.New(sha256.New, []byte(c.secret))
	_, err := h.Write([]byte(content))
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

// req 发起请求
func (c *Client) req(method, address string, params map[string]string) (map[string]interface{}, error) {
	address = c.host + address
	u, err := url.Parse(address)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}
	host := u.Hostname()
	path := u.EscapedPath()

	compute := map[string]string{
		"AccessKeyId":      c.key,
		"SignatureMethod":  "HmacSHA256",
		"SignatureVersion": "2",
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05"),
	}

	var ctype, signature string
	var reader io.Reader
	switch strings.ToUpper(method) {
	case "GET":
		for k, v := range params {
			compute[k] = v
		}
		ctype = "application/x-www-form-urlencoded"
	default:
		ctype = "application/json"

		bs, err := json.Marshal(params)
		if err != nil {
			return nil, errors.Wrap(err, util.FuncName())
		}
		reader = bytes.NewBuffer(bs)
	}

	query := querystring(compute)
	signature, err = c.sign(method + "\n" + host + "\n" + path + "\n" + query)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}
	// huobi get parameters must be passing by querystring
	address += "?" + query + "&Signature=" + url.QueryEscape(signature)

	client := &http.Client{Timeout: time.Duration(time.Second * 3)}

	req, err := http.NewRequest(method, address, reader)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}
	req.Header.Set("Content-Type", ctype)

	var retry int
t:
	resp, err := client.Do(req)
	if err != nil {
		if err, ok := err.(net.Error); (ok && err.Timeout()) ||
			strings.Contains(err.Error(), "connection reset by peer") {
			if retry++; retry < 3 {
				goto t
			}
		}
		return nil, errors.Wrap(err, util.FuncName())
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println(err)
		}
	}()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}
	if err := handle(bs, err); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(bs, &m)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	return m, nil
}

// decode convert an arbitrary map[string]interface{} into a Go structure.
func decode(m map[string]interface{}, i interface{}) error {
	decoder, err := mapstructure.NewDecoder(
		&mapstructure.DecoderConfig{
			WeaklyTypedInput: true,
			Result:           i,
		})
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	err = decoder.Decode(m)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}
	return nil
}
