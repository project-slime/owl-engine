package engine

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"owl-engine/pkg/client/database"
	"owl-engine/pkg/client/influxdb"
	"owl-engine/pkg/client/redis"
	"owl-engine/pkg/config"
	"owl-engine/pkg/service/v0/calculate"
	"owl-engine/pkg/util/signals"
	"owl-engine/pkg/xlogs"
	"owl-engine/router"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var (
	confType   string
	configFile string
	ServerCmd  = &cobra.Command{
		Use:     "server",
		Short:   "Start Engine Server",
		Example: "owl-engine server -t [apollo|file] [-c conf/config.yaml]",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// 加载配置
			switch strings.ToLower(confType) {
			case "file":
				return config.LoadFromFile(configFile)
			case "apollo":
				return config.LoadFromApollo()
			default:
				return config.LoadFromFile(configFile)
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(signals.SetupSignalHandler())
		},
	}
)

// 命令行参数解析
func init() {
	ServerCmd.PersistentFlags().StringVarP(&confType, "type", "t", "file", "Specifies how the configuration is loaded. eg: apollo or file")
	ServerCmd.PersistentFlags().StringVarP(&configFile, "conf", "c", "conf/config.yaml", "specify the configuration file")
}

// 初始化客户端库
func setup(conf *config.ServerRunOptions) {
	// 初始化日志
	xlogs.Log = xlogs.New(
		xlogs.WithLogDir(conf.LoggerOptions.Dir),
		xlogs.WithLogName(conf.LoggerOptions.Name),
		xlogs.WithLogLevel(conf.LoggerOptions.Level),
		xlogs.WithLogFormat(conf.LoggerOptions.Format),
		xlogs.WithLogCaller(conf.LoggerOptions.AddCaller),
		xlogs.WithLoggerMaxSize(conf.LoggerOptions.MaxSize),
		xlogs.WithLoggerMaxAge(conf.LoggerOptions.MaxAge),
		xlogs.WithLoggerMaxBackup(conf.LoggerOptions.MaxBackup),
		xlogs.WithLoggerInterval(conf.LoggerOptions.Interval),
		xlogs.WithLoggerQueue(conf.LoggerOptions.Queue),
		xlogs.WithLoggerQueueSleep(conf.LoggerOptions.QueueSleep), // 单位: 毫秒
		xlogs.WithLoggerDebug(conf.LoggerOptions.Debug),
		xlogs.WithLoggerAsync(conf.LoggerOptions.Async),
		xlogs.WithLoggerCompress(conf.LoggerOptions.Compress),
		xlogs.WithLoggerEncoder()).Build()

	// 数据库
	database.Setup(conf.MySQLOptions)
	influxdb.Setup(conf.InfluxDBOptions)
	redis.Setup(conf.RedisOptions)
}

func taskGoroutine(stopCh <-chan struct{}, wg *sync.WaitGroup) {
	// 	在 golang 中, 为保证不会出现 goroutine 泄漏:
	// 	采用 "如果 goroutine 负责创建 goroutine，它也负责确保它可以停止 goroutine"
	// 	这个约定有助于确保你的程序在组合和扩展时可以扩展
	// 	我们如何确保 goroutine 能够被停止，可以根据 goroutine 的类型和用途而有所不同,
	// 	但是它 们所有这些都是建立在完成 channel传递的基础上的
	wg.Add(3)

	go calculate.Math(stopCh, wg)   // 数学规则
	go calculate.Logger(stopCh, wg) // 日志规则
	go calculate.Warn(stopCh, wg)   // 提醒
}

func run(stopCh <-chan struct{}) error {
	// 获取配置
	conf := config.Get()
	setup(conf)

	var wg sync.WaitGroup
	taskGoroutine(stopCh, &wg)

	// 注意： 日志刷盘和资源释放
	defer func() {
		xlogs.Flush()
		redis.RedisClient.Close()
	}()

	// 指定部署环境
	if strings.Compare(conf.ServerOptions.Mode, "prod") == 0 {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化路由
	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", conf.ServerOptions.Bind, conf.ServerOptions.Port),
		Handler: router.InitRouter(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				xlogs.Errorf("http listening %s:%d error, %s", conf.ServerOptions.Bind, conf.ServerOptions.Port, err.Error())
				xlogs.Flush()
				os.Exit(1)
			}
		}
	}()

	// 打印启动日志
	if strings.Compare(confType, "file") == 0 {
		xlogs.Infof("loading configuration file %s success", configFile)
	} else {
		xlogs.Info("loading configuration from apollo success")
	}
	xlogs.Info(fmt.Sprintf("http listening address %s:%d success", conf.ServerOptions.Bind, conf.ServerOptions.Port))

	<-stopCh
	wg.Wait()

	_ = httpServer.Shutdown(ctx)
	xlogs.Info("Server graceful shutdown success")

	return nil
}

func Execute() error {
	return ServerCmd.Execute()
}
