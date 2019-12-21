package corn

// Logger 日志接口
type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
}
