package corn

import (
	"testing"
	"time"
)

func Test_Parse(t *testing.T) {
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
				{"* 4 15 2 1 * ", &TimeSchedule{last{}, 0xFFFFFFFFFFFFFFF, 0x10, 0x8000, 0x4, 0x2, 0x7F, time.Local}},
				{"5 * 15 2 1 * ", &TimeSchedule{last{}, 0x20, 0xFFFFFFFFFFFFFFF, 0x8000, 0x4, 0x2, 0x7F, time.Local}},
				{"5 4 * 2 1 * ", &TimeSchedule{last{}, 0x20, 0x10, 0xFFFFFF, 0x4, 0x2, 0x7F, time.Local}},
				{"5 4 15 * 1 * ", &TimeSchedule{last{}, 0x20, 0x10, 0x8000, 0xFFFFFFFE, 0x2, 0x7F, time.Local}},
				{"5 4 15 2 * * ", &TimeSchedule{last{}, 0x20, 0x10, 0x8000, 0x4, 0x1FFE, 0x7F, time.Local}},
			},
		},

		// ',' 测试
		{
			"',' 测试",
			[]struct {
				expr string
				want *TimeSchedule
			}{
				{"5,6 4 15 2 1 * ", &TimeSchedule{last{}, 0x60, 0x10, 0x8000, 0x4, 0x2, 0x7F, time.Local}},
				{"5 4,5,24 15 2 1 * ", &TimeSchedule{last{}, 0x20, 0x1000030, 0x8000, 0x4, 0x2, 0x7F, time.Local}},
				{"5 4 15,22 2 1 * ", &TimeSchedule{last{}, 0x20, 0x10, 0x408000, 0x4, 0x2, 0x7F, time.Local}},
				{"5 4 15 2,12 1 * ", &TimeSchedule{last{}, 0x20, 0x10, 0x8000, 0x1004, 0x2, 0x7F, time.Local}},
				{"5 4 15 2 1,5 * ", &TimeSchedule{last{}, 0x20, 0x10, 0x8000, 0x4, 0x22, 0x7F, time.Local}},
			},
		},

		// '-' 测试
		{
			"'-' 测试",
			[]struct {
				expr string
				want *TimeSchedule
			}{
				{"0-59 4 15 2 1 * ", &TimeSchedule{last{}, 0xFFFFFFFFFFFFFFF, 0x10, 0x8000, 0x4, 0x2, 0x7F, time.Local}},
				{"5 0-59 15 2 1 * ", &TimeSchedule{last{}, 0x20, 0xFFFFFFFFFFFFFFF, 0x8000, 0x4, 0x2, 0x7F, time.Local}},
				{"5 4 0-23 2 1 * ", &TimeSchedule{last{}, 0x20, 0x10, 0xFFFFFF, 0x4, 0x2, 0x7F, time.Local}},
				{"5 4 15 1-31 1 * ", &TimeSchedule{last{}, 0x20, 0x10, 0x8000, 0xFFFFFFFE, 0x2, 0x7F, time.Local}},
				{"5 4 15 2 1-12 * ", &TimeSchedule{last{}, 0x20, 0x10, 0x8000, 0x4, 0x1FFE, 0x7F, time.Local}},
				{"5 4 15 2 1 0-6 ", &TimeSchedule{last{}, 0x20, 0x10, 0x8000, 0x4, 0x2, 0x7F, time.Local}},
			},
		},

		// '/' 测试
		{
			"'/' 测试",
			[]struct {
				expr string
				want *TimeSchedule
			}{
				{"*/4 4 15 2 1 * ", &TimeSchedule{last{}, 0x111111111111111, 0x10, 0x8000, 0x4, 0x2, 0x7F, time.Local}},
				{"5 */4 15 2 1 * ", &TimeSchedule{last{}, 0x20, 0x111111111111111, 0x8000, 0x4, 0x2, 0x7F, time.Local}},
				{"5 4 */4 2 1 * ", &TimeSchedule{last{}, 0x20, 0x10, 0x111111, 0x4, 0x2, 0x7F, time.Local}},
				{"5 4 15 */4 1 * ", &TimeSchedule{last{}, 0x20, 0x10, 0x8000, 0x11111110, 0x2, 0x7F, time.Local}},
				{"5 4 15 2 */4 * ", &TimeSchedule{last{}, 0x20, 0x10, 0x8000, 0x4, 0x1110, 0x7F, time.Local}},
				{"5 4 15 2 1 */3 ", &TimeSchedule{last{}, 0x20, 0x10, 0x8000, 0x4, 0x2, 0x49, time.Local}},
			},
		},

		// 复杂字符串
		{
			"'*'、','、'-'、'/'组合测试",
			[]struct {
				expr string
				want *TimeSchedule
			}{
				{"0-12/4 * * * * * ", &TimeSchedule{last{}, 0x1111, 0xFFFFFFFFFFFFFFF, 0xFFFFFF, 0xFFFFFFFE, 0x1FFE, 0x7F, time.Local}},
				{"0-12/4,20-30/5 * * * * * ", &TimeSchedule{last{}, 0x42101111, 0xFFFFFFFFFFFFFFF, 0xFFFFFF, 0xFFFFFFFE, 0x1FFE, 0x7F, time.Local}},
				// ','分割范围交叉覆盖
				{"0-12/4,20-30/5,25-45/3 * * * * * ", &TimeSchedule{last{}, 0x924D2101111, 0xFFFFFFFFFFFFFFF, 0xFFFFFF, 0xFFFFFFFE, 0x1FFE, 0x7F, time.Local}},
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
				{"0-62 0-60 0-24 0-32 0-63 0-63 ", &TimeSchedule{last{}, 0xFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFF, 0xFFFFFF, 0xFFFFFFFE, 0x1FFE, 0x7F, time.Local}},
				// 无效取值
				{"60-63 60-63 24-63 32-63 13-63 7-63 ", &TimeSchedule{last{}, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, time.Local}},
				// 不合法参数
				{"64 64 64 64 64 64 ", &TimeSchedule{last{}, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, time.Local}},
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
						_ts.day != val.want.day || _ts.month != val.want.month {
						t.Errorf("expr: %s, get: %+v, want: %+v", val.expr, _ts, val.want)
					}
				default:
					t.Errorf("不知道的调度题类型")
				}

			}

		})
	}
}
