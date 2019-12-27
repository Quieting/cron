package corn

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// 测试任务管理是否准确执行
func Test_Process(t *testing.T) {
	c := New()
	go c.Run()

	now := time.Now()
	taskWithTime := func() error {
		fmt.Printf("我只执行一次\n")
		return nil
	}

	c.AddWithTime(now.Add(10*time.Second), taskWithTime)

	times := 0
	var s sync.Mutex
	taskWithCorn := func() error {
		s.Lock()
		defer s.Unlock()
		times++
		fmt.Printf("我执行了 %d 次\n", times)
		return nil
	}

	_, _, sec := now.Clock()
	corn := fmt.Sprintf("%d,%d,%d * * * * *", (sec+1)%60, (sec+5)%60, (sec+10)%60)
	c.AddWithCorn(corn, taskWithCorn)

	time.Sleep(20 * time.Second)
	c.Stop()

}
