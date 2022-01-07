package xlogs

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	logger := New(
		WithLogName("app.log"),
		WithLogLevel("info"),
		WithLogCaller(true),
		WithLoggerMaxSize(1),   // 单位: MB, 为测试文件分割, 设置为 1MB, 默认为 512MB
		WithLoggerMaxAge(7),    // 单位: 天
		WithLoggerMaxBackup(7), // 单位: 份数
		WithLoggerInterval(24), // 单位: 小时
		WithLoggerQueue(true),
		WithLoggerQueueSleep(100), // 单位: 毫秒
		WithLoggerDebug(false),
		WithLoggerAsync(true),
		WithLoggerCompress(true),
		WithLoggerEncoder()).Build()

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
