/*
Option模式的优缺点
优点
	支持传递多个参数，并且在参数个数、类型发生变化时保持兼容性
	任意顺序传递参数
	支持默认值
	方便拓展
缺点
	增加许多function，成本增大
	参数不太复杂时，尽量少用
*/

package xlogs

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type OptionFunc func(config *LoggerConfig)

// LoggerConfig 基础配置
type LoggerConfig struct {
	Dir           string                 `json:"dir" yaml:"dir"`                 // 日志目录
	Name          string                 `json:"name" yaml:"name"`               // 日志文件
	Level         string                 `json:"level" yaml:"level"`             // 日志级别
	Format        string                 `json:"format" yaml:"format"`           // 日志输出格式
	AddCaller     bool                   `json:"add_caller" yaml:"add_caller"`   // 是否添加调用者信息
	CallerSkip    int                    `json:"caller_skip" yaml:"caller_skip"` // 调用者的最大层级
	MaxSize       int                    `json:"max_size" yaml:"max_size"`       // 日志文件的最大尺寸
	MaxAge        int                    `json:"max_age" yaml:"max_age"`         // 日志文件存在的最大生命周期
	MaxBackup     int                    `json:"max_backup" yaml:"max_backup"`   // 日志保留的最大份数
	Interval      time.Duration          `json:"interval" yaml:"interval"`       // 日志磁盘刷新间隔
	Async         bool                   `json:"async" yaml:"async"`             // 是否异步刷盘
	Queue         bool                   `json:"queue" yaml:"queue"`             // 是否启用队列,进行日志内容缓存
	QueueSleep    time.Duration          `json:"queue_sleep" yaml:"queue_sleep"` // 唤醒队列间隔
	Fields        []zap.Field            `json:"fields" yaml:"fields"`           // 自定义字段
	Core          zapcore.Core           `json:"core" yaml:"core"`
	Debug         bool                   `json:"debug" yaml:"debug"`                   // 是否启用debug模式
	EncoderConfig *zapcore.EncoderConfig `json:"encoder_config" yaml:"encoder_config"` // 日志输出格式化器
	Compress      bool                   `json:"compress" yaml:"compress"`             // 是否压缩
}

// New 初始化配置
func New(fns ...OptionFunc) *LoggerConfig {
	// 先进行默认初始化配置
	var config = &LoggerConfig{
		Dir:        "./logs",
		Name:       "app.log",
		Level:      "info",
		Format:     "json",
		AddCaller:  true,
		CallerSkip: 2,
		MaxSize:    512, // 单位: MB
		MaxAge:     7,   // 单位: 天
		MaxBackup:  7,   // 单位: 份
		Interval:   24,  // 单位: 小时
		Async:      true,
		Queue:      true,
		QueueSleep: 100, // 单位: 毫秒
		Debug:      false,
		Compress:   false,
	}

	// 执行初始化配置参数
	for _, fn := range fns {
		fn(config)
	}

	return config
}

func (l LoggerConfig) Build() *Logger {
	if l.EncoderConfig == nil {
		l.EncoderConfig = EncoderConfig()
	}

	if l.Debug {
		l.EncoderConfig.EncodeLevel = DebugEncoderLevel
	}

	logger := newLogger(&l)

	return logger
}

// Filename 日志路径和文件名称
func (l *LoggerConfig) Filename() string {
	if strings.Compare(l.Dir, "") == 0 {
		path, _ := filepath.Abs(os.Args[0])
		l.Dir = filepath.Join(path)
	}
	if strings.Compare(l.Name, "") == 0 {
		l.Name = "app.log"
	}

	return fmt.Sprintf("%s/%s", l.Dir, l.Name)
}

// WithLogDir 设置日志存储目录
func WithLogDir(dir string) OptionFunc {
	return func(config *LoggerConfig) {
		if strings.Compare(dir, "") == 0 {
			dir, _ = filepath.Abs(os.Args[0])
		}
		config.Dir = dir
	}
}

// WithLogName 设置日志文件名
func WithLogName(name string) OptionFunc {
	return func(config *LoggerConfig) {
		if strings.Compare(name, "") == 0 {
			name = "app.log"
		}

		config.Name = name
	}
}

// WithLogLevel 设置日志级别
func WithLogLevel(level string) OptionFunc {
	return func(config *LoggerConfig) {
		if strings.Compare(level, "") == 0 {
			level = "debug"
		}

		config.Level = level
	}
}

func WithLogFormat(format string) OptionFunc {
	return func(config *LoggerConfig) {
		ff := "json"

		switch strings.ToLower(format) {
		case "json":
			ff = "json"
		case "common":
			ff = "console"
		default:
			ff = "json"
		}

		config.Format = ff
	}
}

func WithLogCaller(caller bool) OptionFunc {
	return func(config *LoggerConfig) {
		config.AddCaller = caller
	}
}

func WithLoggCallerSkip(skip int) OptionFunc {
	return func(config *LoggerConfig) {
		if skip == 0 {
			skip = 2
		}

		config.CallerSkip = skip
	}
}

func WithLoggerMaxSize(size int) OptionFunc {
	return func(config *LoggerConfig) {
		if size == 0 {
			// 默认为 512MB
			size = 512 * 1024 * 1024
		}
		config.MaxSize = size
	}
}

func WithLoggerMaxAge(age int) OptionFunc {
	return func(config *LoggerConfig) {
		if age == 0 {
			age = 7
		}
		config.MaxAge = age
	}
}

func WithLoggerMaxBackup(backup int) OptionFunc {
	return func(config *LoggerConfig) {
		if backup == 0 {
			backup = 7
		}

		config.MaxBackup = backup
	}
}

func WithLoggerInterval(interval int) OptionFunc {
	return func(config *LoggerConfig) {
		if interval == 0 || interval > 24 {
			interval = 24
		}

		config.Interval = time.Duration(interval) * time.Hour
	}
}

func WithLoggerAsync(async bool) OptionFunc {
	return func(config *LoggerConfig) {
		config.Async = async
	}
}

func WithLoggerQueue(queue bool) OptionFunc {
	return func(config *LoggerConfig) {
		config.Queue = queue
	}
}

func WithLoggerQueueSleep(sleep int) OptionFunc {
	return func(config *LoggerConfig) {
		if sleep == 0 || sleep > 1000 {
			sleep = 1000
		}

		config.QueueSleep = time.Duration(sleep) * time.Millisecond
	}
}

func WithLoggerDebug(debug bool) OptionFunc {
	return func(config *LoggerConfig) {
		config.Debug = debug
	}
}

// WithLoggerFields 主要为分布式微服务进行标识的自定义输出字段
func WithLoggerFields(field map[string]string) OptionFunc {
	return func(config *LoggerConfig) {
		// 使用反射, 获取自定义的字段和值
		keyValue := reflect.ValueOf(field)
		for _, v := range keyValue.MapKeys() {
			key := v.String()
			config.Fields = append(config.Fields, String(key, field[key]))
		}
	}
}

func WithLoggerEncoder() OptionFunc {
	return func(config *LoggerConfig) {
		config.EncoderConfig = EncoderConfig()
	}
}

func WithLoggerCompress(compress bool) OptionFunc {
	return func(config *LoggerConfig) {
		config.Compress = compress
	}
}
