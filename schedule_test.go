package corn

import (
	"fmt"
	"math"
	"testing"
	"time"
)

// String 格式化输出
func (t *TimeSchedule) String() string {
	return fmt.Sprintf("{sYear: %d, second: %x, min: %x, hour: %x, day: %x, month: %x, weekDay: %x, year: %x}",
		t.sYear, t.second, t.min, t.hour, t.day, t.month, t.weekDay, t.year)
}

func Test_Parse(t *testing.T) {
	year := time.Now().Year()
	data := []struct {
		name string

		d []struct {
			expr string
			want *TimeSchedule
		}
	}{
		// '*' 测试
		{
			"'*' 测试",
			[]struct {
				expr string
				want *TimeSchedule
			}{
				{"* 4 15 2 1 * ", &TimeSchedule{year, 0xFFFFFFFFFFFFFFF, 0x10, 0x8000, 0x4, 0x2, 0x7F, math.MaxUint64, time.Local}},
				{"5 * 15 2 1 * ", &TimeSchedule{year, 0x20, 0xFFFFFFFFFFFFFFF, 0x8000, 0x4, 0x2, 0x7F, math.MaxUint64, time.Local}},
				{"5 4 * 2 1 * ", &TimeSchedule{year, 0x20, 0x10, 0xFFFFFF, 0x4, 0x2, 0x7F, math.MaxUint64, time.Local}},
				{"5 4 15 * 1 * ", &TimeSchedule{year, 0x20, 0x10, 0x8000, 0xFFFFFFFE, 0x2, 0x7F, math.MaxUint64, time.Local}},
				{"5 4 15 2 * * ", &TimeSchedule{year, 0x20, 0x10, 0x8000, 0x4, 0x1FFE, 0x7F, math.MaxUint64, time.Local}},
			},
		},

		// ',' 测试
		{
			"',' 测试",
			[]struct {
				expr string
				want *TimeSchedule
			}{
				{"5,6 4 15 2 1 * ", &TimeSchedule{year, 0x60, 0x10, 0x8000, 0x4, 0x2, 0x7F, math.MaxUint64, time.Local}},
				{"5 4,5,24 15 2 1 * ", &TimeSchedule{year, 0x20, 0x1000030, 0x8000, 0x4, 0x2, 0x7F, math.MaxUint64, time.Local}},
				{"5 4 15,22 2 1 * ", &TimeSchedule{year, 0x20, 0x10, 0x408000, 0x4, 0x2, 0x7F, math.MaxUint64, time.Local}},
				{"5 4 15 2,12 1 * ", &TimeSchedule{year, 0x20, 0x10, 0x8000, 0x1004, 0x2, 0x7F, math.MaxUint64, time.Local}},
				{"5 4 15 2 1,5 * ", &TimeSchedule{year, 0x20, 0x10, 0x8000, 0x4, 0x22, 0x7F, math.MaxUint64, time.Local}},
			},
		},

		// '-' 测试
		{
			"'-' 测试",
			[]struct {
				expr string
				want *TimeSchedule
			}{
				{"0-59 4 15 2 1 * ", &TimeSchedule{year, 0xFFFFFFFFFFFFFFF, 0x10, 0x8000, 0x4, 0x2, 0x7F, math.MaxUint64, time.Local}},
				{"5 0-59 15 2 1 * ", &TimeSchedule{year, 0x20, 0xFFFFFFFFFFFFFFF, 0x8000, 0x4, 0x2, 0x7F, math.MaxUint64, time.Local}},
				{"5 4 0-23 2 1 * ", &TimeSchedule{year, 0x20, 0x10, 0xFFFFFF, 0x4, 0x2, 0x7F, math.MaxUint64, time.Local}},
				{"5 4 15 1-31 1 * ", &TimeSchedule{year, 0x20, 0x10, 0x8000, 0xFFFFFFFE, 0x2, 0x7F, math.MaxUint64, time.Local}},
				{"5 4 15 2 1-12 * ", &TimeSchedule{year, 0x20, 0x10, 0x8000, 0x4, 0x1FFE, 0x7F, math.MaxUint64, time.Local}},
				{"5 4 15 2 1 0-6 ", &TimeSchedule{year, 0x20, 0x10, 0x8000, 0x4, 0x2, 0x7F, math.MaxUint64, time.Local}},
			},
		},

		// '/' 测试
		{
			"'/' 测试",
			[]struct {
				expr string
				want *TimeSchedule
			}{
				{"*/4 4 15 2 1 * ", &TimeSchedule{year, 0x111111111111111, 0x10, 0x8000, 0x4, 0x2, 0x7F, math.MaxUint64, time.Local}},
				{"5 */4 15 2 1 * ", &TimeSchedule{year, 0x20, 0x111111111111111, 0x8000, 0x4, 0x2, 0x7F, math.MaxUint64, time.Local}},
				{"5 4 */4 2 1 * ", &TimeSchedule{year, 0x20, 0x10, 0x111111, 0x4, 0x2, 0x7F, math.MaxUint64, time.Local}},
				{"5 4 15 */4 1 * ", &TimeSchedule{year, 0x20, 0x10, 0x8000, 0x11111110, 0x2, 0x7F, math.MaxUint64, time.Local}},
				{"5 4 15 2 */4 * ", &TimeSchedule{year, 0x20, 0x10, 0x8000, 0x4, 0x1110, 0x7F, math.MaxUint64, time.Local}},
				{"5 4 15 2 1 */3 ", &TimeSchedule{year, 0x20, 0x10, 0x8000, 0x4, 0x2, 0x49, math.MaxUint64, time.Local}},
			},
		},

		// 复杂字符串
		{
			"'*'、','、'-'、'/'组合测试",
			[]struct {
				expr string
				want *TimeSchedule
			}{
				{"0-12/4 * * * * * ", &TimeSchedule{year, 0x1111, 0xFFFFFFFFFFFFFFF, 0xFFFFFF, 0xFFFFFFFE, 0x1FFE, 0x7F, math.MaxUint64, time.Local}},
				{"0-12/4,20-30/5 * * * * * ", &TimeSchedule{year, 0x42101111, 0xFFFFFFFFFFFFFFF, 0xFFFFFF, 0xFFFFFFFE, 0x1FFE, 0x7F, math.MaxUint64, time.Local}},
				// ','分割范围交叉覆盖
				{"0-12/4,20-30/5,25-45/3 * * * * * ", &TimeSchedule{year, 0x924D2101111, 0xFFFFFFFFFFFFFFF, 0xFFFFFF, 0xFFFFFFFE, 0x1FFE, 0x7F, math.MaxUint64, time.Local}},
			},
		},

		// 非常规字符串
		{
			"容错恢复测试",
			[]struct {
				expr string
				want *TimeSchedule
			}{
				// 取值溢出
				{"0-62 0-60 0-24 0-32 0-63 0-63 ", &TimeSchedule{year, 0xFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFF, 0xFFFFFF, 0xFFFFFFFE, 0x1FFE, 0x7F, math.MaxUint64, time.Local}},
				// 无效取值
				{"60-63 60-63 24-63 32-63 13-63 7-63 ", &TimeSchedule{year, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, math.MaxUint64, time.Local}},
				// 不合法参数
				{"64 64 64 64 64 64 ", &TimeSchedule{year, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, math.MaxUint64, time.Local}},
			},
		},
	}

	for _, p := range data {
		t.Run(p.name, func(t *testing.T) {
			for _, val := range p.d {
				ts, err := Parse(val.expr)
				if err != nil {
					t.Error(err)
				}

				switch _ts := ts.(type) {
				case *TimeSchedule:
					if _ts.second != val.want.second || _ts.min != val.want.min || _ts.hour != val.want.hour ||
						_ts.day != val.want.day || _ts.month != val.want.month || _ts.year != val.want.year || _ts.sYear != val.want.sYear {
						t.Errorf("expr: %s, get: %+v, want: %+v", val.expr, _ts, val.want)
					}
				default:
					t.Errorf("不知道的调度题类型")
				}

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
		{"测试空时间", time.Time{}, TimeSchedule{1, 1, 1, 1, 2, 2, 2, 0, time.Local}},
		{"测试空的月份", time.Date(2019, 5, 20, 0, 0, 0, 0, time.Local), TimeSchedule{2019, 1, 1, 1, 1 << 20, 1 << 5, 1 << 1, 0, time.Local}},
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
		// 基础测试
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

		// 跨时区测试
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
