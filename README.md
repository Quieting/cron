# corn
任务调度,继续完善中

# Install
```
  go get github.com/Quiteing/corn
```

# Get Starting

```
package main

import "github.com/Quiteing/corn"

func main() {
  cn := corn.New()
  
  // 开启
  go cn.Run()
}
```