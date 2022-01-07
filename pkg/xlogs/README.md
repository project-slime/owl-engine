## 日志库设计

### 设计思路

* 具备特点
    * 日志格式
        * json
        * 普通字符串格式化
    * 时间和刷新频率
        * 自定义刷盘间隔
        * 何时滚动日志: 按照日志尺寸和时间
    * 安全性
        * 写日志安全加锁
        * 双 buffer 机制, 避免每一次push一条消息时就要唤醒buffer内存空间
        * 设计一个并发安全的高性能buffer

### demo 使用

```go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/StalkerOne/xlogs"
)

func main() {
	logger := xlogs.New(
		xlogs.WithLogName("app.log"),
		xlogs.WithLogLevel("info"),
		xlogs.WithLogCaller(true),
		xlogs.WithLoggerMaxSize(1),   // 单位: MB, 为测试文件分割, 设置为 1MB, 默认为 512MB
		xlogs.WithLoggerMaxAge(7),    // 单位: 天
		xlogs.WithLoggerMaxBackup(7), // 单位: 份数
		xlogs.WithLoggerInterval(24), // 单位: 小时
		xlogs.WithLoggerQueue(true),
		xlogs.WithLoggerQueueSleep(100), // 单位: 毫秒
		xlogs.WithLoggerDebug(false),
		xlogs.WithLoggerAsync(true),
		xlogs.WithLoggerCompress(true),
		xlogs.WithLoggerEncoder()).Build()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			current := time.Now().Format("2006-01-02 15:04:05")
			fmt.Printf("current datetime is: %s", current)
			logger.Infof("current datetime is: %s.", current)
			logger.Error("Insufficient host memory")
			fmt.Println(time.Now())
		case <-shutdown:
			ticker.Stop()
			_ = logger.Flush()
			return
		}
	}

}
```