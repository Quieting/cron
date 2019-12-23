package corn

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// 错误
var (
	ErrInvialParam = errors.New("无效的参数")
)

// 各时间单位取值范围
const (
	RangeSec     = 0xFFFFFFFFFFFFFFF
	RangeMin     = 0xFFFFFFFFFFFFFFF
	RangeHour    = 0xFFFFFF
	RangeDay     = 0xFFFFFFFE
	RangeMonth   = 0x1FFE
	RangeWeekDay = 0x7F
)

// Scheduler 调度器
type Scheduler interface {
	Next(t time.Time) time.Time
}

// TimeSchedule 时间调度
type TimeSchedule struct {
	// 起始年
	sYear int

	// 秒、分、小时、日期、月份、星期、年
	second, min, hour, day, month, weekDay, year uint64

	loc *time.Location
}

// Max 调度器最后一次执行时间
func (t *TimeSchedule) Max() time.Time {
	year := uint64(63)
	month := uint64(12)
	day := uint64(31)
	hour := uint64(23)
	min := uint64(59)
	sec := uint64(59)

	for {
		// 获取年
		year = findBitBack(t.year, year, 0)
		if year < 0 {
			return time.Time{}
		}

		// 获取月
		month = findBitBack(t.month, month, 1)
		if month < 1 {
			return time.Time{}
		}

		// 获取日
		day = findBitBack(t.day, day, 1)
		if day < 1 {
			return time.Time{}
		}

		_t := time.Date(t.sYear+int(year), time.Month(month), int(day), 0, 0, 0, 0, time.Local)
		_year, _month, _day := _t.Date()
		if _year != int(year) || _month != time.Month(month) || _day != int(day) || (1<<uint64(_t.Weekday()))&t.weekDay == 0 {
			day--
			continue
		}

		// 获取时
		hour = findBitBack(t.hour, hour, 0)
		if hour < 0 {
			return time.Time{}
		}

		// 获取分
		min = findBitBack(t.min, min, 0)
		if min < 0 {
			return time.Time{}
		}

		// 获取秒
		sec = findBitBack(t.second, sec, 0)
		if sec < 0 {
			return time.Time{}
		}

		break
	}

	return time.Date(t.sYear+int(year), time.Month(month), int(day), int(hour), int(min), int(sec), 0, time.Local)
}

// findBit 从低位向高位查找直到指为 1 的 bit 位(0-63)
func findBit(n, start, end uint64) uint64 {
	for start < end+1 {
		if (n & (1 << start)) > 0 {
			break
		}
		start++
	}
	return start
}

// findBitBack 从高位向低位查找直到指为 1 的 bit 位(0-63)
func findBitBack(n, start, end uint64) uint64 {
	for start > end-1 {
		if (n & (1 << start)) > 0 {
			break
		}
		start--
	}
	return start
}

// Next 符合 TimeSchedule 的下个时间
func (t *TimeSchedule) Next(_time time.Time) time.Time {
	// 原始时区
	oriLoc := _time.Location()

	// 统一时区
	_time = _time.In(t.loc)

	year, month, day := _time.Date()
	hour, min, sec := _time.Clock()

	look := func(_t time.Time) (ts time.Time, b bool) {
		// 找到符合的年
		_y := _t.Year() - t.sYear
		if _y < 0 {
			_y = 0
		}
		y := findBit(t.year, uint64(_y), 63)
		if y > 63 {
			return time.Time{}, false
		}

		year = t.sYear + int(y)

		// 找到符合要求的月
		m := findBit(t.month, uint64(month), 12)
		if m > 12 {
			year++
			month = 1
			day = 1
			hour = 0
			min = 0
			sec = 0
			return time.Date(year, 1, 1, 0, 0, 0, 0, _t.Location()), false
		}
		if m != uint64(month) {
			day = 1
			hour = 0
			min = 0
			sec = 0
		}

		month = time.Month(m)

		// 找到符合要求的天
		for d := uint16(day); d < 33; d++ {
			if (t.day & (1 << d)) == 0 {
				if d == 32 {
					month++
					day = 1
					hour = 0
					min = 0
					sec = 0
					return time.Date(year, month, 1, 0, 0, 0, 0, _t.Location()), false
				}
				continue
			}

			// 验证新生成的时间是否跨月、跨年, 检测星期是否符合
			_t := time.Date(year, month, int(d), 0, 0, 0, 0, _t.Location())
			_year, _month, _day := _t.Date()
			if _year != year || _month != month || _day != int(d) || (1<<uint64(_t.Weekday()))&t.weekDay == 0 {
				continue
			}

			if d != uint16(day) {
				hour = 0
				min = 0
				sec = 0
			}

			day = int(d)
			break
		}

		// 找到符合要求的时
		h := findBit(t.hour, uint64(hour), 23)
		if h > 23 {
			day++
			hour = 0
			min = 0
			sec = 0
			return time.Date(year, month, day, 0, 0, 0, 0, _t.Location()), false
		}

		if h != uint64(hour) {
			min = 0
			sec = 0
		}

		hour = int(h)

		// 找到符合要求的分
		mm := findBit(t.min, uint64(min), 59)
		if mm > 59 {
			hour++
			min = 0
			sec = 0
			return time.Date(year, month, day, hour, 0, 0, 0, _t.Location()), false
		}

		if mm != uint64(min) {
			sec = 0
		}

		min = int(mm)

		// 找到符合要求的秒
		s := findBit(t.second, uint64(sec), 59)
		if s > 59 {
			min++
			sec = 0
			return time.Date(year, month, day, hour, min, 0, 0, _t.Location()), false
		}
		sec = int(s)

		return time.Date(year, month, day, hour, min, sec, 0, _t.Location()), true
	}

	for {
		var b bool
		_time, b = look(_time)
		if b {
			return _time.In(oriLoc)
		}
		if _time.IsZero() {
			return _time
		}

		if _time.Year()-year > 5 {
			return time.Time{}
		}
	}
}

// ts 无效数据修正
func (t *TimeSchedule) amend() {
	t.second = t.second & RangeSec
	t.min = t.min & RangeMin
	t.hour = t.hour & RangeHour
	t.day = t.day & RangeDay
	t.month = t.month & RangeMonth
	t.weekDay = t.weekDay & RangeWeekDay
}

// Time2TimeSchedule 将时间转换成调度器
func Time2TimeSchedule(_time time.Time) *TimeSchedule {
	t := new(TimeSchedule)
	hour, min, sec := _time.Clock()
	t.second = 1 << uint64(sec)
	t.min = 1 << uint64(min)
	t.hour = 1 << uint64(hour)
	year, mon, day := _time.Date()
	t.sYear = year
	t.year = 1 << uint64(year-t.sYear)
	t.day = 1 << uint64(day)
	t.month = 1 << uint64(mon)
	t.weekDay = 1 << uint64(_time.Weekday())
	return t
}

// Parse 按照 linux crontab 格式解析时间字符串
//
// spec: 以空格分割
// 前6个字段分别表示：
// .-------------- 秒钟：0-59
// | .------------ 分钟：0-59
// | | . --------- 小时：0-23
// | | | .-------- 日期：1-31
// | | | | .------ 月份：1-12
// | | | | | .---- 星期：0-6（0 表示周日）
// | | | | | | .-- 年:  0-63(如果起始年是2019, 0表示2019年, 10表示2029年)
// | | | | | | |
// * * * * * * *
/////////////////////////////////////////////////////////
// 还可以用一些特殊符号：
// *:   表示任何时刻
// ,:   表示分割，如第三段里：2,4，表示 2 点和 4 点执行
// －:  表示一个段，如第三端里： 1-5，就表示 1 到 5 点
// /n:  表示每个n的单位执行一次，如第三段里，*/1, 就表示每隔 1 个小时执行一次命令。也可以写成1-23/1.
/////////////////////////////////////////////////////////
// 举例如下:
//  0/30 * * * * * *                      每 30 秒 执行
//  0 5,15 5 * * * *　　                   5:5, 05:15 执行
//  0 0-10 17 * * * *                     17:00 到 17:10 毎隔 1 分钟 执行
//  0 2 8-20/3 * * * *　　　　　　          8:02,11:02,14:02,17:02,20:02 执行
func Parse(spec string) (s Scheduler, err error) {
	// 1.按空格分割字符串获取时间参数
	params := strings.Fields(spec)

	l := len(params)
	for i := 0; i < 7-l; i++ {
		params = append(params, "*")
	}

	ts := new(TimeSchedule)
	ts.sYear = time.Now().Year()

	f := func(str string) (_time uint64, err error) {
		commas := strings.Split(str, ",")
		for _, comma := range commas {
			var _t uint64
			_t, err = parse(comma)
			if err != nil {
				return
			}
			_time |= _t
		}

		return
	}

	// 2.解析参数
	ts.second, err = f(params[0])
	if err != nil {
		return
	}

	ts.min, err = f(params[1])
	if err != nil {
		return
	}

	ts.hour, err = f(params[2])
	if err != nil {
		return
	}

	ts.day, err = f(params[3])
	if err != nil {
		return
	}

	ts.month, err = f(params[4])
	if err != nil {
		return
	}

	ts.weekDay, err = f(params[5])
	if err != nil {
		return
	}

	ts.year, err = f(params[6])
	if err != nil {
		return
	}

	// 3.修正不合法数据
	ts.amend()

	s = ts
	return
}

// parse 解析如下格式字符串:
// number | number "-" number [ "/" number ] | *[ "/" number]
func parse(expr string) (_time uint64, err error) {
	var (
		frequency  = uint64(0)
		start, end = uint64(0), uint64(0)
		slash      = strings.Split(expr, "/")
	)

	// 解析'/'之后的字符串
	switch len(slash) {
	case 1:
		frequency = 1
	case 2:
		frequency, err = strconv.ParseUint(slash[1], 10, 64)
		if err != nil || frequency < 0 {
			err = ErrInvialParam
			return
		}
	default:
		err = ErrInvialParam
		return
	}

	// 解析 '/' 之前的字符串
	hyphen := strings.Split(slash[0], "-")
	switch len(hyphen) {
	case 1:
		if hyphen[0] == "*" {
			start, end = 0, 63
			break
		}
		start, err = strconv.ParseUint(hyphen[0], 10, 64)
		if err != nil {
			err = ErrInvialParam
			return
		}
		end = start
	case 2:
		start, err = strconv.ParseUint(hyphen[0], 10, 64)
		if err != nil {
			err = ErrInvialParam
			return
		}
		end, err = strconv.ParseUint(hyphen[1], 10, 64)
		if err != nil {
			err = ErrInvialParam
			return
		}
	default:
		err = ErrInvialParam
		return
	}

	if start > 63 {
		start = 64
	}
	if end > 63 {
		end = 63
	}

	for i := start; i < end+1; i += frequency {
		_time = bitSet(_time, i, 1)
	}

	return
}

// 设置 b 第 index 位的值为 val(从右向左计算)
// index: 0 - 63
// val: 0、1
func bitSet(b, index, val uint64) uint64 {
	switch val {
	case 0:
		b = b & (^(1 << index))
	case 1:
		b = b | 1<<index
	}
	return b
}
