package corn

import (
	"errors"
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

// Scheduler 时间调度器，根据当前时间返回下一个符合定义的时间
type Scheduler interface {
	Laster
	// Next 下一个符合要求的时间,返回零时则表示没有找到符合要求的时间
	// 当返回零时，大于 t 的时间将一直返回零时
	Next(t time.Time) time.Time
}

// TimeSchedule 时间调度器
type TimeSchedule struct {
	last
	// 秒、分、小时、日期、月份、星期、年
	second, min, hour, day, month, weekDay uint64

	loc *time.Location
}

// Laster 最后执行时间
type Laster interface {
	// Last 如果返回时间是零时则表示时间调度器没有限制将一直执行
	Last() time.Time
}

// last Laster 空实现
type last struct{}

// Last 调度器最后一次执行时间
func (last) Last() time.Time {
	return time.Time{}
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
	// 时间进1到秒
	_time = _time.Add(1*time.Second - time.Duration(_time.Nanosecond())*time.Nanosecond)

	// 原始时区
	oriLoc := _time.Location()

	// 统一时区
	_time = _time.In(t.loc)

	year, month, day := _time.Date()
	hour, min, sec := _time.Clock()

	look := func(_t time.Time) (ts time.Time, b bool) {

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
	_, mon, day := _time.Date()
	t.day = 1 << uint64(day)
	t.month = 1 << uint64(mon)
	t.weekDay = 1 << uint64(_time.Weekday())
	return t
}

type convert func(start uint64, end uint64, frequency uint64) []uint64

func convertDate(start uint64, end uint64, frequency uint64) []uint64 {
	var _time uint64
	for i := start; i < end; i += frequency {
		_time = bitSet(_time, i, 1)
	}
	return []uint64{_time}
}

func convertYear(start uint64, end uint64, frequency uint64) []uint64 {
	var years []uint64
	for i := start; i < end; i += frequency {
		years = append(years, i)
	}
	return years
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

// DurationSchedule 时间调度器(定长时间执行,比如每72小时执行一次)
type DurationSchedule struct {
	last

	// 起始时间
	start time.Time

	// 间隔执行时间
	frequency time.Duration
}

// Next 临近 t 的下一次执行时机
func (d *DurationSchedule) Next(t time.Time) time.Time {
	dur := t.Sub(d.start)
	if quo := dur % d.frequency; quo != 0 {
		return d.start.Add((dur/d.frequency + 1) * d.frequency)

	}
	return t
}

var _ Scheduler = new(DurationSchedule)

// FixSchedule 时间调度器(固定时间执行,只执行一次)
type FixSchedule struct {
	rTime time.Time
}

// Next 临近 t 的下一次执行时机
func (f *FixSchedule) Next(t time.Time) time.Time {
	return f.rTime
}

// Last 最后一次执行时间
func (f *FixSchedule) Last() time.Time {
	return f.rTime
}

var _ Scheduler = new(FixSchedule)
