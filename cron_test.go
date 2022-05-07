package corn

import (
	"testing"
	"time"
)

func Test_cron_Add(t *testing.T) {
	type args struct {
		scheduler Scheduler
		job       Job
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "示例: 未运行过程中添加 job",
			args: args{
				scheduler: &FixSchedule{time.Now().Add(10 * time.Second)},
				job: JobFunc(func() error {
					return nil
				}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := defaultCorner()
			if got := c.Add(tt.args.scheduler, tt.args.job); got != tt.want {
				t.Errorf("Add() = %v, want %v", got, tt.want)
			}
		})
	}
}
