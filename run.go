package corn

import (
	"fmt"
	"sort"
	"time"

	"github.com/bwmarrin/snowflake"
)

// Jober 任务管理器管理的对象
type Jober interface {
	Run() bool
	Next() time.Time
	SetNext()
	Name() string
}

type jobs []Jober

func (j jobs) Len() int {
	return len(j)
}
func (j jobs) Less(i, k int) bool {
	return j[i].Next().Before(j[k].Next())
}

func (j jobs) Swap(i, k int) {
	j[i], j[k] = j[k], j[i]
}

// Corn 任务调度器
type Corn struct {
	jobs []Jober // 待运行的任务

	node *snowflake.Node // 用于生成任务唯一标识
	log  Logger

	a chan Jober    // 添加任务
	d chan string   // 删除任务
	s chan struct{} // 停止任务管理
}

// New 初始化 Corn 对象
func New(log Logger) *Corn {
	corn := new(Corn)
	corn.log = log
	corn.a = make(chan Jober)
	corn.d = make(chan string)
	corn.s = make(chan struct{})

	// 2019-11-11 0:00:00
	snowflake.Epoch = 1573401600456
	corn.node, _ = snowflake.NewNode(1)
	if log == nil {

	}
	return corn
}

// RunFunc 任务执行体
type RunFunc func() error

// AddWithTime 添加执行一次的任务
// name: 任务标识(唯一),可通过此标识删除某不需要执行的任务
func (c *Corn) AddWithTime(j time.Time, f RunFunc) (name string) {
	name = c.node.Generate().Base58()
	job := new(job)
	job.name = name
	job.f = f
	job.next = j
	c.a <- job
	return
}

// AddWithCorn 添加重复执行定时任务
func (c *Corn) AddWithCorn(corn string, f RunFunc) (name string, err error) {
	name = c.node.Generate().Base58()
	job := new(job)
	job.name = name
	job.f = f
	job.ts, err = Parse(corn)
	job.SetNext()
	if err != nil {
		return
	}

	c.a <- job
	return
}

// Add 添加任务
func (c *Corn) Add(f RunFunc, s Scheduler) (name string) {
	name = c.node.Generate().Base58()
	job := new(job)
	job.name = name
	job.f = f
	job.s = s
	job.SetNext()

	c.a <- job
	return
}

// Delete 删除任务
func (c *Corn) Delete(name string) {
	c.d <- name
	return
}

// Run 运行任务
func (c *Corn) Run() {
	for {
		now := time.Now()
		sort.Sort(jobs(c.jobs))

		var effc time.Time // 下一次任务执行时间
		for _, j := range c.jobs {
			if next := j.Next(); next.After(now) {
				effc = next
				break
			}
		}
		if effc.IsZero() {
			effc = now.Add(10 * time.Minute)
		}

		select {
		case <-time.NewTicker(effc.Sub(now)).C:
			for _, j := range c.jobs {
				if j.Next().Equal(effc) {
					go func(j Jober) {
						if j.Run() {
							c.Delete(j.Name())
							return
						}
					}(j)
					j.SetNext()
					continue
				}
				break
			}
		case j := <-c.a: // 添加任务
			c.jobs = append(c.jobs, j)
		case name := <-c.d: // 删除任务
			for index, j := range c.jobs {

				if j.Name() != name {
					continue
				}
				c.jobs = append(c.jobs[:index], c.jobs[index+1:]...)
				break
			}
		}
	}
}

type job struct {
	f        RunFunc
	ts       *TimeSchedule // 重复任务专用
	name     string        // 任务名字
	next     time.Time     // 下一次执行时间
	status   string        // 1.执行失败 2.执行中
	times    int           // 执行次数
	errInfos []errInfo     // 执行失败的错误信息
	log      Logger
	s        Scheduler
}

// Run 返回当前任务是否已结束
// true: 已结束
// false: 未结束需要继续执行
func (j *job) Run() bool {
	err := j.f()
	if err != nil {
		j.log.Error(err)
		j.errInfos = append(j.errInfos, errInfo{
			msg: err.Error(),
			t:   time.Now(),
		})
		return false
	}
	if j.ts == nil {
		return true
	}
	return false
}

func (j *job) SetNext() {
	if j.next.IsZero() {
		j.next = time.Now()
	}
	// 将 j.next 转换成整秒
	j.next = j.next.Add(1*time.Second - time.Duration(j.next.Nanosecond())*time.Nanosecond)
	if j.ts != nil {
		j.next = j.ts.Next(j.next)
		fmt.Printf("next:%v\n", j.next)
		return
	}
	switch j.times {
	case 0:
		j.next = j.next.Add(5 * time.Second)
	case 1:
		j.next = j.next.Add(15 * time.Second)
	case 2:
		j.next = j.next.Add(30 * time.Second)
	case 3:
		j.next = j.next.Add(1 * time.Minute)
	case 4:
		j.next = j.next.Add(5 * time.Minute)
	}

	j.times++
}

func (j *job) Next() time.Time {
	return j.next
}

func (j *job) Name() string {
	return j.name
}

type errInfo struct {
	msg string
	t   time.Time
}
