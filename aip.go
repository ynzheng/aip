package main

import (
	"log"
	"time"

	"github.com/modood/aip/db"
	"github.com/modood/aip/huobi"
	"github.com/modood/aip/plan"
	"github.com/modood/aip/util"

	"github.com/pkg/errors"
	"github.com/robfig/cron"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	name = "aip"
	desc = "aip - automatic investment plan for digital currency"
)

var (
	errUnkownPeriod = errors.New("unknown period")
)

var cmd = &cobra.Command{
	Use:  name,
	Long: desc,
	RunE: execute,

	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	tloc, err := time.LoadLocation("Asia/Chongqing")
	if err != nil {
		tloc = time.FixedZone("CST", 3600*8)
	}
	time.Local = tloc

	if err := initFlags(); err != nil {
		log.Fatalln(errors.Wrap(err, util.FuncName()))
	}
}

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalln(errors.Wrap(err, util.FuncName()))
	}

	select {}
}

func initFlags() error {
	// TODO 展示排序
	flags := cmd.Flags()

	viper.AutomaticEnv()
	viper.SetEnvPrefix(name)

	flags.String("dbfile", "/var/opt/aip.sqlite3", "sqlite3 data file path")
	if err := viper.BindPFlag("dbfile", flags.Lookup("dbfile")); err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	flags.String("apikey", "", "huobi api key")
	if err := viper.BindPFlag("apikey", flags.Lookup("apikey")); err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	flags.String("apisecret", "", "huobi api secret")
	if err := viper.BindPFlag("apisecret", flags.Lookup("apisecret")); err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	flags.String("apihost", "https://api.huobi.pro", "huobi api host")
	if err := viper.BindPFlag("apihost", flags.Lookup("apihost")); err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	flags.String("symbol", "btcusdt", "symbol name")
	if err := viper.BindPFlag("symbol", flags.Lookup("symbol")); err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	flags.Float64("amount", 0, "investment amount")
	if err := viper.BindPFlag("amount", flags.Lookup("amount")); err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	flags.String("period", "daily", "period of automatic investment.\navailable: daily, weekly and monthly")
	if err := viper.BindPFlag("period", flags.Lookup("period")); err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	return nil
}

func execute(cmd *cobra.Command, args []string) error {
	var (
		c   *huobi.Client
		p   plan.Period
		pl  plan.Plan
		err error
	)

	dbfile := viper.GetString("dbfile")
	apikey := viper.GetString("apikey")
	apisecret := viper.GetString("apisecret")
	apihost := viper.GetString("apihost")
	symbol := viper.GetString("symbol")
	amount := viper.GetFloat64("amount")
	period := viper.GetString("period")

	// TODO 参数校验

	// 初始化数据库
	if err = db.Init(dbfile); err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	// 创建火币客户端
	c, err = huobi.NewClient(apihost, apikey, apisecret)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	// 创建定投周期
	switch period {
	case "daily":
		p, err = plan.NewDaily(0, 0, 0)
	case "weekly":
		p, err = plan.NewWeekly(time.Monday, 0, 0, 0)
	case "monthly":
		p, err = plan.NewMonthly(1, 0, 0, 0)
	}
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	// 创建定投计划
	pl, err = plan.New(symbol, amount, p, c)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	// 执行定投计划
	if err = run(pl); err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	return nil
}

func run(p plan.Plan) error {
	invest := cron.New()
	if err := invest.AddFunc(p.Period().Schedule(), func() {
		if err := p.Invest(); err != nil {
			log.Println(err)
		}
	}); err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	monitor := cron.New()
	if err := monitor.AddFunc("0 0 * * * *", func() {
		if err := p.Monitor(); err != nil {
			log.Println(err)
		}
	}); err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	invest.Start()
	monitor.Start()

	return nil
}
