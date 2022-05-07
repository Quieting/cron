package corn

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/bwmarrin/snowflake"
)

const (
	def     = 0
	stop    = 1
	running = 2
)

// Corner 管理定时任务管理器，负责 Job 的添加，删除，执行和停止
// 同一个 Job 同一时间只会有一个实例在执行
type Corner interface {
	// Add 添加 Job，返回 Job 唯一标识
	// 相同的 Job 反复调用也会被添加成两个任务
	// 任何情况下都可以调用(包括运行过程中)，并发安全
	Add(scheduler Scheduler, job Job) string

	// Delete 删除任务，
	// 如果 Job 不存在不会返回错误
	// 如果 Job 正在执行中，将完成本次执行后删除，方法返回将不需要等待执行结束
	// 任何情况下都可以调用(包括运行过程中)，并发安全
	Delete(id string)

	// Run 已运行重复调用不会产生任何影响
	Run()

	// Stop 停止运行，未运行情况下调用不会产生任何影响
	// 已运行 Job 执行结束后停止
	Stop()
}

// Cron
// todo：封装接口对外提供功能
type Cron struct {
	Corner
}

// CronOption 扩展 Cron 功能使用
type CronOption func(c *Cron)

func NewCorn(opts ...CronOption) *Cron {
	c := &Cron{}
	c.Corner = defaultCorner()
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// AddWithTime 添加执行一次的任务
// name: 任务标识(唯一),可通过此标识删除某不需要执行的任务
func (c *Cron) AddWithTime(t time.Time, f Func) error {
	c.Add(&FixSchedule{t}, JobFunc(f))
	return nil
}

// AddWithCorn 添加重复执行定时任务
func (c *Cron) AddWithCorn(expr string, f Func) error {
	scheduler, err := Parse(expr)
	if err != nil {
		return err
	}
	c.Add(scheduler, JobFunc(f))
	return nil
}

func defaultCorner() Corner {
	node, _ := snowflake.NewNode(1)
	return &cron{
		del:  make(chan string),
		stop: make(chan struct{}),
		add:  make(chan *entity),
		work: make(chan string),
		wg:   &sync.WaitGroup{},
		jobs: make(map[string]*entity),
		node: node,
	}
}

// todo：负责运行过程中的 Job 调度
type cron struct {
	del  chan string
	stop chan struct{}
	add  chan *entity
	work chan string

	state int64

	wg   *sync.WaitGroup
	jobs map[string]*entity
	node *snowflake.Node
}

type entity struct {
	work chan string
	id   string
	Job
	Scheduler
}

func (c *cron) Add(scheduler Scheduler, job Job) string {
	id := c.node.Generate().String()
	e := &entity{
		id:        id,
		Job:       job,
		Scheduler: scheduler,
	}
	select {
	case c.add <- e:
	default:
		c.jobs[id] = e
	}

	return id
}

func (c *cron) Delete(id string) {
	c.del <- id
	return
}

// Run
// todo: 目前并发调用 Run 和 Add 方法会造成 c.jobs 的并发读写，后续优化
func (c *cron) Run() {
	if !atomic.CompareAndSwapInt64(&c.state, def, running) {
		return
	}
	go func() {
		for _, job := range c.jobs {
			c.add <- job
		}
	}()

	for {
		select {
		case id := <-c.del:
			delete(c.jobs, id)
		case e := <-c.add:
			c.addJob(e)
		case id := <-c.work:
			e, ok := c.jobs[id]
			if !ok {
				continue
			}

			c.do(e)
		case <-c.stop:
			c.wg.Done()
			atomic.SwapInt64(&c.state, def)
			return
		}
	}
}

func (c *cron) Stop() {
	if atomic.CompareAndSwapInt64(&c.state, running, stop) {
		c.stop <- struct{}{}
	}
}

func (c *cron) addJob(e *entity) {
	now := time.Now()
	next := e.Next(now)
	if next.IsZero() || next.Before(now) {
		return
	}
	c.jobs[e.id] = e
	go func(id string) {
		time.Sleep(next.Sub(next))
		c.work <- e.id
	}(e.id)
}

func (c *cron) do(e *entity) {
	c.wg.Add(1)
	go func(wg *sync.WaitGroup, e *entity) {
		defer wg.Done()
		e.Run()
		c.addJob(e)
	}(c.wg, e)
}
