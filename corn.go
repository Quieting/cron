package corn

import (
	"sync"

	"github.com/bwmarrin/snowflake"
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

type Corn struct {
	mux sync.RWMutex

	jobs map[string]Job
	node *snowflake.Node
}

func (c *Corn) Add(scheduler Scheduler, job Job) string {
	c.mux.Lock()
	defer c.mux.Unlock()
	id := node.Generate().String()
	c.jobs[id] = entity{
		Job:       job,
		Scheduler: scheduler,
	}
	return id
}

func (c *Corn) Delete(id string) {
	panic("implement me")
}

func (c *Corn) Run() {
	panic("implement me")
}

func (c *Corn) Stop() {
	panic("implement me")
}
