package config

import (
	"owl-engine/pkg/util/reflectutils"

	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerOptions struct {
	Dir           string                 `json:"dir" yaml:"dir"`                // 日志目录
	Name          string                 `json:"name" yaml:"name"`              // 日志文件
	Level         string                 `json:"level" yaml:"level"`            // 日志级别
	Format        string                 `json:"format" yaml:"format"`          // 日志输出格式
	AddCaller     bool                   `json:"add_caller" yaml:"addCaller"`   // 是否添加调用者信息
	CallerSkip    int                    `json:"caller_skip" yaml:"callerSkip"` // 调用者的最大层级
	MaxSize       int                    `json:"max_size" yaml:"maxSize"`       // 日志文件的最大尺寸
	MaxAge        int                    `json:"max_age" yaml:"maxAge"`         // 日志文件存在的最大生命周期
	MaxBackup     int                    `json:"max_backup" yaml:"maxBackup"`   // 日志保留的最大份数
	Interval      int                    `json:"interval" yaml:"interval"`      // 日志磁盘刷新间隔
	Async         bool                   `json:"async" yaml:"async"`            // 是否异步刷盘
	Queue         bool                   `json:"queue" yaml:"queue"`            // 是否启用队列,进行日志内容缓存
	QueueSleep    int                    `json:"queue_sleep" yaml:"queueSleep"` // 唤醒队列间隔
	Fields        []zap.Field            `json:"fields" yaml:"fields"`          // 自定义字段
	Core          zapcore.Core           `json:"core" yaml:"core"`
	Debug         bool                   `json:"debug" yaml:"debug"`                   // 是否启用debug模式
	EncoderConfig *zapcore.EncoderConfig `json:"encoder_config" yaml:"encoder_config"` // 日志输出格式化器
	Compress      bool                   `json:"compress" yaml:"compress"`             // 是否压缩
}

func NewLoggerOptions() *LoggerOptions {
	return &LoggerOptions{
		Dir:        "./logs",
		Name:       "owl-engine.log",
		Level:      "info",
		Format:     "json",
		AddCaller:  true,
		CallerSkip: 2,
		MaxSize:    128,
		MaxAge:     7,
		MaxBackup:  7,
		Interval:   24,
		Async:      true,
		Queue:      true,
		QueueSleep: 100,
		Debug:      false,
		Compress:   true,
	}
}

func (l *LoggerOptions) Validate() []error {
	errors := make([]error, 0)

	return errors
}

func (l *LoggerOptions) ApplyTo(options *LoggerOptions) {
	reflectutils.Override(options, l)
}

func (l *LoggerOptions) AddFlags(fs *pflag.FlagSet) {

	fs.StringVar(&l.Dir, "log-dir", l.Dir, "Specify log storage directory")

	fs.StringVar(&l.Name, "log-name", l.Name, "Specify log name")

	fs.StringVar(&l.Level, "log-level", l.Level, "Specify log level. eg: info, debug, warning, error, fatal")

	fs.StringVar(&l.Format, "log-format", l.Format, "Specify log output format")
}
