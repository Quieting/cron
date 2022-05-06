package corn

import (
	"fmt"
	"testing"
	"time"
)

// String 格式化输出
func (t *TimeSchedule) String() string {
	return fmt.Sprintf("{second: %x, min: %x, hour: %x, day: %x, month: %x, weekDay: %x}",
		t.second, t.min, t.hour, t.day, t.month, t.weekDay)
}

func Test_Time2TimeSchedule(t *testing.T) {
	type paramTime2TS struct {
		name  string
		_time time.Time
		want  TimeSchedule
	}

	data := []paramTime2TS{
		{"测试空时间", time.Time{}, TimeSchedule{last{}, 1, 1, 1, 2, 2, 2, time.Local}},
		{"测试空的月份", time.Date(2019, 5, 20, 0, 0, 0, 0, time.Local), TimeSchedule{last{}, 1, 1, 1, 1 << 20, 1 << 5, 1 << 1, time.Local}},
	}

	for _, p := range data {
		t.Run(p.name, func(t *testing.T) {
			ts := Time2TimeSchedule(p._time)
			if ts.second != p.want.second || ts.min != p.want.min || ts.hour != p.want.hour ||
				ts.day != p.want.day || ts.month != p.want.month {
				t.Errorf("get: %+v, want: %+v", ts, p.want)
			}
		})
	}
}

func Test_TimeScheduleNext(t *testing.T) {
	type paramTime2TS struct {
		name string
		expr string
		now  time.Time
		next time.Time
	}

	data := []paramTime2TS{
		{"测试当前月份", "0 20 5 * * *", time.Date(2019, 5, 20, 0, 0, 0, 0, time.Local), time.Date(2019, 5, 20, 5, 20, 0, 0, time.Local)},
		{"测试跨年", "0 20 5 28,31 4 *", time.Date(2019, 5, 28, 7, 0, 0, 0, time.Local), time.Date(2020, 4, 28, 5, 20, 0, 0, time.Local)},
		{"测试跨月", "0 20 5 28,31 * *", time.Date(2019, 2, 28, 7, 0, 0, 0, time.Local), time.Date(2019, 3, 28, 5, 20, 0, 0, time.Local)},
		{"测试跨天", "0 20 5 28,31 * *", time.Date(2019, 3, 29, 7, 0, 0, 0, time.Local), time.Date(2019, 3, 31, 5, 20, 0, 0, time.Local)},
		{"测试跨时", "0 20 5 28,31 * *", time.Date(2019, 2, 28, 4, 0, 0, 0, time.Local), time.Date(2019, 2, 28, 5, 20, 0, 0, time.Local)},
		{"测试跨分", "0 20 5 28,31 * *", time.Date(2019, 2, 28, 5, 19, 0, 0, time.Local), time.Date(2019, 2, 28, 5, 20, 0, 0, time.Local)},
		{"测试跨秒", "0 20 5 28,31 * *", time.Date(2019, 2, 28, 5, 19, 23, 0, time.Local), time.Date(2019, 2, 28, 5, 20, 0, 0, time.Local)},
		{"测试星期天", "* * 3 * * 0", time.Date(2019, 11, 20, 1, 19, 23, 0, time.Local), time.Date(2019, 11, 24, 3, 0, 0, 0, time.Local)},
		{"测试分", "*/5 10 * * *", time.Date(2019, 2, 28, 5, 10, 23, 0, time.Local), time.Date(2019, 2, 28, 5, 10, 25, 0, time.Local)},
		{"测试无符合条件的时间", "* * * 32 3", time.Date(2019, 2, 28, 5, 10, 23, 0, time.Local), time.Time{}},
	}

	for _, p := range data {
		t.Run(p.name, func(t *testing.T) {
			ts, _ := Parse(p.expr)
			if get := ts.Next(p.now); !get.Equal(p.next) {
				t.Errorf("ts: %v, want: %s, get: %s", ts, p.next.Format("2006-01-02 15:04:05"), get.Format("2006-01-02 15:04:05"))
			}
		})
	}
}

func Test_TimeScheduleMax(t *testing.T) {
	type paramTime2TS struct {
		name string
		expr string
		max  time.Time
	}

	data := []paramTime2TS{
		{"执行一次", "5 4 15 2 1 *", time.Date(2019, 1, 2, 15, 4, 5, 0, time.Local)},
	}

	for _, p := range data {
		t.Run(p.name, func(t *testing.T) {
			ts, _ := Parse(p.expr)
			if get := ts.Last(); !get.IsZero() {
				t.Errorf("ts: %v, get: %s", ts, get.Format("2006-01-02 15:04:05"))
			}
		})
	}
}

func Benchmark_Range(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parse("0-31/4")
	}
}
func Benchmark_Parse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Parse("0-31/4,40-50/2 20 15 * *")
	}
}

func Benchmark_Next(b *testing.B) {
	ts, _ := Parse("0 20 5 28,31 * *")
	now := time.Date(2020, 4, 28, 5, 20, 0, 0, time.Local)
	for i := 0; i < b.N; i++ {
		ts.Next(now)
	}
}
