package plan

import (
	"fmt"
	"time"

	"github.com/modood/aip/util"

	"github.com/pkg/errors"
)

// Period 周期接口定义
type Period interface {
	// Schedule 获取 cron 表达式
	Schedule() string
}

type datetime struct {
	hour   uint8
	minute uint8
	second uint8
}

// Daily 按日
type Daily struct {
	datetime
}

// NewDaily 创建按日周期
func NewDaily(hour, minute, second uint8) (Period, error) {
	if err := validateTime(hour, minute, second); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	return &Daily{
		datetime: datetime{
			hour:   hour,
			minute: minute,
			second: second,
		},
	}, nil
}

// Schedule 获取 cron 表达式
func (d *Daily) Schedule() string {
	return fmt.Sprintf("%d %d %d * * *", d.second, d.minute, d.hour)
}

// Weekly 按周
type Weekly struct {
	datetime
	weekday time.Weekday
}

// NewWeekly 创建按周周期
func NewWeekly(weekday time.Weekday, hour, minute, second uint8) (Period, error) {
	if err := validateTime(hour, minute, second); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	return &Weekly{
		weekday: weekday,
		datetime: datetime{
			hour:   hour,
			minute: minute,
			second: second,
		},
	}, nil
}

// Schedule 获取 cron 表达式
func (w *Weekly) Schedule() string {
	return fmt.Sprintf("%d %d %d * * %d", w.second, w.minute, w.hour, w.weekday)
}

// Monthly 按月
type Monthly struct {
	datetime
	day uint8
}

// NewMonthly 创建按月周期
func NewMonthly(day, hour, minute, second uint8) (Period, error) {
	if err := validateTime(hour, minute, second); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	return &Monthly{
		day: day,
		datetime: datetime{
			hour:   hour,
			minute: minute,
			second: second,
		},
	}, nil
}

// Schedule 获取 cron 表达式
func (m *Monthly) Schedule() string {
	return fmt.Sprintf("%d %d %d %d * *", m.second, m.minute, m.hour, m.day)
}

func validateTime(hour, minute, second uint8) error {
	var err error
	if err = mustBetween(hour, 0, 23); err != nil {
		return errors.Wrap(err, "invalid hour")
	}
	if err = mustBetween(minute, 0, 59); err != nil {
		return errors.Wrap(err, "invalid minute")
	}
	if err = mustBetween(second, 0, 59); err != nil {
		return errors.Wrap(err, "invalid second")
	}
	return nil
}

func mustBetween(value, from, to uint8) error {
	if value < from || value > to {
		return fmt.Errorf("a valid value must between %d and %d (inclusive) but got %d", from, to, value)
	}
	return nil
}
