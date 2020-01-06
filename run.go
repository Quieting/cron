package corn

import (
	"sort"
	"sync"
	"time"

	"github.com/bwmarrin/snowflake"
)

var node *snowflake.Node

func init() {
	snowflake.Epoch = time.Now().Unix()
	node, _ = snowflake.NewNode(1)
}

// Jober 任务管理器管理的对象
type Jober interface {
	// Run 返回当前对象是否需要再执行
	// true: 继续执行, false: 已结束,无需再执行
	Run() bool
	// 下次执行时间
	Next() time.Time
	SetNext(time.Time)
	// Name 身份标识,删除不再执行的任务时使用
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

	log Logger

	add  chan Jober    // 添加任务
	del  chan string   // 删除任务
	stop chan struct{} // 停止任务管理

	group sync.WaitGroup // 停止任务时保证任务执行成功
}

// New 初始化 Corn 对象
func New() *Corn {
	corn := new(Corn)
	corn.add = make(chan Jober)
	corn.del = make(chan string)
	corn.stop = make(chan struct{})
	return corn
}

// SetLog 设置日志
func (c *Corn) SetLog(log Logger) {
	c.log = log
}

// RunFunc 任务执行体
type RunFunc func() error

// AddWithTime 添加执行一次的任务
// name: 任务标识(唯一),可通过此标识删除某不需要执行的任务
func (c *Corn) AddWithTime(rTime time.Time, f RunFunc) {
	c.Add(f, &FixSchedule{rTime})
	return
}

// AddWithCorn 添加重复执行定时任务
func (c *Corn) AddWithCorn(corn string, f RunFunc) error {
	sche, err := Parse(corn)
	if err != nil {
		return err
	}
	c.Add(f, sche)
	return nil
}

// Add 添加任务
func (c *Corn) Add(f RunFunc, s Scheduler) {
	entry := new(Entry)
	entry.ID = node.Generate()
	entry.Schedule = s
	entry.Func = f

	c.add <- entry
}

// Delete 删除任务
func (c *Corn) Delete(name string) {
	c.del <- name
	return
}

// Stop 结束任务
func (c *Corn) Stop() {
	c.stop <- struct{}{}
	return
}

// Run 运行任务
func (c *Corn) Run() {
	c.log.Info("定时任务已启动...")
	for {
		now := time.Now()
		for _, j := range c.jobs {
			j.SetNext(now)
		}
		sort.Sort(jobs(c.jobs))

		var effc time.Time // 下一次任务执行时间
		for _, j := range c.jobs {
			next := j.Next()
			if next.After(now) {
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
					c.group.Add(1)
					go func(j Jober) {
						if !j.Run() {
							c.del <- j.Name()
						}
						c.group.Done()
						return
					}(j)
				}
			}
		case j := <-c.add: // 添加任务
			c.jobs = append(c.jobs, j)
		case name := <-c.del:
			for index, val := range c.jobs {
				if val.Name() != name {
					continue
				}
				c.jobs = append(c.jobs[:index], c.jobs[index+1:]...)
				c.log.Info("删除任务: %s", val.Name())
				break
			}
		case <-c.stop:
			c.group.Wait()
			c.log.Info("定时任务已结束")
			return
		}
	}
}

// Entry 执行单位
type Entry struct {
	ID       snowflake.ID
	Func     RunFunc
	Schedule Scheduler // 时间调度器
	ErrInfos []errInfo // 执行失败的错误信息
	Log      Logger

	m    sync.Mutex
	next time.Time
}

// Run 返回当前对象是否需要再执行
// true: 继续执行
// false: 已结束,无需再执行
func (e *Entry) Run() bool {
	err := e.Func()
	now := time.Now()
	if err != nil {
		e.Log.Error(err)
		e.ErrInfos = append(e.ErrInfos, errInfo{
			msg: err.Error(),
			t:   now,
		})
		return true
	}

	// 判断是否需要继续执行
	last := e.Schedule.Last()

	if last.IsZero() {
		return true
	}

	if last.Before(now) {
		return false
	}

	return true
}

// SetNext 设置下一次执行时间
func (e *Entry) SetNext(now time.Time) {
	e.m.Lock()
	// 设置下一次执行时间
	e.next = e.Schedule.Next(now)
	e.m.Unlock()
}

// Next 下一次执行时间
func (e *Entry) Next() time.Time {
	e.m.Lock()
	_t := e.next
	e.m.Unlock()
	return _t
}

// Name  唯一标识
func (e *Entry) Name() string {
	return e.ID.String()
}

type errInfo struct {
	msg string
	t   time.Time
}
