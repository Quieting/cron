package corn

// Job 抽象的 Job 接口，保留扩展能力
type Job interface {
	Run() error
}

// Func 向定时器加入新的定时任务时，外部传入的参数，应包装成 Job 使用
type Func func() error

type JobFunc func() error

func (f JobFunc) Run() error {
	return f()
}
