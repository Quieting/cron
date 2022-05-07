package corn

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

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
//
// 还可以用一些特殊符号：
// *:   表示任何时刻
// ,:   表示分割，如第三段里：2,4，表示 2 点和 4 点执行
// －:  表示一个段，如第三端里： 1-5，就表示 1 到 5 点
// /n:  表示每个n的单位执行一次，如第三段里，*/1, 就表示每隔 1 个小时执行一次命令。也可以写成1-23/1.
//
// 举例如下:
//  0/30 * * * * * *                      每 30 秒 执行
//  0 5,15 5 * * * *　　                   5:5, 05:15 执行
//  0 0-10 17 * * * *                     17:00 到 17:10 毎隔 1 分钟 执行
//  0 2 8-20/3 * * * *　　　　　　          8:02,11:02,14:02,17:02,20:02 执行
func Parse(spec string) (s Scheduler, err error) {
	// 1.按空格分割字符串获取时间参数
	params := strings.Fields(spec)

	if l := len(params); l != 6 {
		return nil, fmt.Errorf("需要 6 个参数,只传入 %d 个参数", l)
	}

	ts := new(TimeSchedule)

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

	ts.loc = time.Local

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
