package corn

import (
	"math"
	"testing"
	"time"
)

func Test_Parse(t *testing.T) {
	type paramParse struct {
		name string
		expr string
		want *TimeSchedule
	}

	data := []paramParse{
		{"五个参数", "0-31/4,32-40/2 20 15 * *", &TimeSchedule{2019, 0x15511111111, 0x100000, 0x8000, 0xFFFFFFFE, 0x1FFE, 0x7F, math.MaxUint64}},
		{"六个参数", "0-31/4,32-40/2 20 15 * ", &TimeSchedule{2019, 0x15511111111, 0x100000, 0x8000, 0xFFFFFFFE, 0x1FFE, 0x7F, math.MaxUint64}},
		{"不合格的参数", "0-31/4,32-40/2 20 15 100 ", &TimeSchedule{2019, 0x15511111111, 0x100000, 0x8000, 0xFFFFFFFE, 0x1FFE, 0x7F, math.MaxUint64}},
		{"测试当前月份", "0 20 5 * * *", &TimeSchedule{2019, 0x1, 0x100000, 0x20, 0xFFFFFFFE, 0x1FFE, 0x7F, math.MaxUint64}},
	}

	t.Logf("(0x1FFE & (1 << 12)) : %d", (0x1FFE & (1 << 12)))

	for _, p := range data {
		t.Run(p.name, func(t *testing.T) {
			ts, err := Parse(p.expr)
			if err != nil {
				t.Error(err)
			}

			if ts.second != p.want.second || ts.min != p.want.min || ts.hour != p.want.hour ||
				ts.day != p.want.day || ts.month != p.want.month || ts.year != p.want.year || ts.sYear != p.want.sYear {
				t.Errorf("get: %+v, want: %+v", ts, p.want)
			}
		})
	}
}
func Test_bitSet(t *testing.T) {
	type paramBitSet struct {
		name          string
		b, index, val uint64
		want          uint64
	}

	data := []paramBitSet{
		{"第 3 位值设为1", 0, 2, 1, 4},
		{"第 3 位值设为0", 2, 2, 0, 2},
		{"第 10 位值设为0", 0xFFFFFFFF, 9, 0, 0xFFFFFDFF},
	}

	for _, p := range data {
		t.Run(p.name, func(t *testing.T) {
			if get := bitSet(p.b, p.index, p.val); get != p.want {
				t.Errorf("want: %x, get: %x", p.want, get)
			}
		})
	}
}

func Test_Range(t *testing.T) {
	type paramRange struct {
		name string
		expr string
		want uint64
	}

	data := []paramRange{
		{"测试星号", "*", math.MaxUint64},
		{"测试频率", "*/4", 0x1111111111111111},
		{"测试破折号", "0-31", 0xFFFFFFFF},
		{"完整测试", "0-31/4", 0x11111111},
	}

	for _, p := range data {
		t.Run(p.name, func(t *testing.T) {
			if get, _ := parse(p.expr); get != p.want {
				t.Errorf("want: %x, get: %x", p.want, get)
			}
		})
	}
}

func Test_Time2TimeSchedule(t *testing.T) {
	type paramTime2TS struct {
		name  string
		_time time.Time
		want  TimeSchedule
	}

	data := []paramTime2TS{
		{"测试空时间", time.Time{}, TimeSchedule{1, 1, 1, 1, 2, 2, 2, 0}},
		{"测试空的月份", time.Date(2019, 5, 20, 0, 0, 0, 0, time.Local), TimeSchedule{2019, 1, 1, 1, 1 << 20, 1 << 5, 1 << 1, 0}},
	}

	for _, p := range data {
		t.Run(p.name, func(t *testing.T) {
			ts := Time2TimeSchedule(p._time)
			if ts.second != p.want.second || ts.min != p.want.min || ts.hour != p.want.hour ||
				ts.day != p.want.day || ts.month != p.want.month || ts.year != p.want.year || ts.sYear != p.want.sYear {
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

func Test_Any(t *testing.T) {
	t.Logf("time.Now().Year(): %d\n", time.Now().Year())
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
