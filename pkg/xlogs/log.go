package xlogs

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	DebugLevel = zap.DebugLevel
	InfoLevel  = zap.InfoLevel
	WarnLevel  = zap.WarnLevel
	ErrorLevel = zap.ErrorLevel
	PanicLevel = zap.PanicLevel
	FatalLevel = zap.FatalLevel
)

type (
	Func   func(string, ...zap.Field)
	Field  = zap.Field
	Level  = zapcore.Level
	Logger struct {
		deSugar *zap.Logger
		level   *zap.AtomicLevel
		config  LoggerConfig
		sugar   *zap.SugaredLogger
	}
)

var (
	String     = zap.String
	Any        = zap.Any
	Int64      = zap.Int64
	Int        = zap.Int
	Int32      = zap.Int32
	Uint       = zap.Uint
	Duration   = zap.Duration
	Durationp  = zap.Durationp
	Object     = zap.Object
	Namespace  = zap.Namespace
	Reflect    = zap.Reflect
	Skip       = zap.Skip()
	ByteString = zap.ByteString
)

func EncoderConfig() *zapcore.EncoderConfig {
	return &zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     timeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// DebugEncoderLevel 日志高亮显示
func DebugEncoderLevel(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var colorize = Red
	switch level {
	case zapcore.DebugLevel:
		colorize = Blue
	case zapcore.InfoLevel:
		colorize = Green
	case zapcore.WarnLevel:
		colorize = Yellow
	case zapcore.ErrorLevel, zapcore.PanicLevel, zapcore.DPanicLevel, zapcore.FatalLevel:
		colorize = Red
	default:
	}

	enc.AppendString(colorize(level.CapitalString()))
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func newLogger(config *LoggerConfig) *Logger {
	zapOptions := make([]zap.Option, 0)
	zapOptions = append(zapOptions, zap.AddStacktrace(zap.DPanicLevel))

	if config.AddCaller {
		zapOptions = append(zapOptions, zap.AddCaller(), zap.AddCallerSkip(config.CallerSkip))
	}

	if len(config.Fields) > 0 {
		zapOptions = append(zapOptions, zap.Fields(config.Fields...))
	}

	var ws zapcore.WriteSyncer
	if config.Debug {
		ws = os.Stdout
	} else {
		// 写文件,并进行轮替和压缩
		ws = zapcore.AddSync(newRotate(config))
	}

	// 异步刷盘
	if config.Async {
		var close CloseFunc
		ws, close = Buffer(ws, defaultBufferSize, defaultFlushInterval)

		// 为其分布式应用进程中的安全退出, 需要进行特别的注销函数
		Register(close)
	}

	level := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	if err := level.UnmarshalText([]byte(config.Level)); err != nil {
		panic(err)
	}

	// 对输出日志进行格式化
	encoder := *config.EncoderConfig
	core := config.Core
	if core == nil {
		// 实现匿名函数
		fn := func() zapcore.Encoder {
			if strings.Compare(config.Format, "console") == 0 {
				return zapcore.NewConsoleEncoder(encoder)
			}
			return zapcore.NewJSONEncoder(encoder)
		}

		core = zapcore.NewCore(fn(), ws, level)
	}

	zapLogger := zap.New(core, zapOptions...)

	return &Logger{
		deSugar: zapLogger,
		level:   &level,
		config:  *config,
		sugar:   zapLogger.Sugar(),
	}
}

// Flush 应用退出后, 执行刷盘
func (logger *Logger) Flush() error {
	return logger.deSugar.Sync()
}

// 实现日志级别接口方法

func (logger *Logger) IsDebugMode() bool {
	return logger.config.Debug
}

func normalizeMessage(msg string) string {
	return fmt.Sprintf("%-32s", msg)
}

// Debug ...
func (logger *Logger) Debug(msg string, fields ...Field) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.deSugar.Debug(msg, fields...)
}

func (logger *Logger) Debugw(msg string, keysAndValues ...interface{}) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.sugar.Debugw(msg, keysAndValues...)
}

func sprintf(template string, args ...interface{}) string {
	msg := template
	if msg == "" && len(args) > 0 {
		msg = fmt.Sprint(args...)
	} else if msg != "" && len(args) > 0 {
		msg = fmt.Sprintf(template, args...)
	}
	return msg
}

func (logger *Logger) StdLog() *log.Logger {
	return zap.NewStdLog(logger.deSugar)
}

func (logger *Logger) Debugf(template string, args ...interface{}) {
	logger.sugar.Debugw(sprintf(template, args...))
}

func (logger *Logger) Info(msg string, fields ...Field) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.deSugar.Info(msg, fields...)
}

func (logger *Logger) Infow(msg string, keysAndValues ...interface{}) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.sugar.Infow(msg, keysAndValues...)
}

func (logger *Logger) Infof(template string, args ...interface{}) {
	logger.sugar.Infof(sprintf(template, args...))
}

func (logger *Logger) Warn(msg string, fields ...Field) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.deSugar.Warn(msg, fields...)
}

func (logger *Logger) Warnw(msg string, keysAndValues ...interface{}) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.sugar.Warnw(msg, keysAndValues...)
}

func (logger *Logger) Warnf(template string, args ...interface{}) {
	logger.sugar.Warnf(sprintf(template, args...))
}

func (logger *Logger) Error(msg string, fields ...Field) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.deSugar.Error(msg, fields...)
}

func (logger *Logger) Errorw(msg string, keysAndValues ...interface{}) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.sugar.Errorw(msg, keysAndValues...)
}

func (logger *Logger) Errorf(template string, args ...interface{}) {
	logger.sugar.Errorf(sprintf(template, args...))
}

func (logger *Logger) Panic(msg string, fields ...Field) {
	if logger.IsDebugMode() {
		panicDetail(msg, fields...)
		msg = normalizeMessage(msg)
	}
	logger.deSugar.Panic(msg, fields...)
}

func (logger *Logger) Panicw(msg string, keysAndValues ...interface{}) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.sugar.Panicw(msg, keysAndValues...)
}

func (logger *Logger) Panicf(template string, args ...interface{}) {
	logger.sugar.Panicf(sprintf(template, args...))
}

func (logger *Logger) DPanic(msg string, fields ...Field) {
	if logger.IsDebugMode() {
		panicDetail(msg, fields...)
		msg = normalizeMessage(msg)
	}
	logger.deSugar.DPanic(msg, fields...)
}

func (logger *Logger) DPanicw(msg string, keysAndValues ...interface{}) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.sugar.DPanicw(msg, keysAndValues...)
}

func (logger *Logger) DPanicf(template string, args ...interface{}) {
	logger.sugar.DPanicf(sprintf(template, args...))
}

func (logger *Logger) Fatal(msg string, fields ...Field) {
	if logger.IsDebugMode() {
		panicDetail(msg, fields...)
		msg = normalizeMessage(msg)
		return
	}
	logger.deSugar.Fatal(msg, fields...)
}

func (logger *Logger) Fatalw(msg string, keysAndValues ...interface{}) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.sugar.Fatalw(msg, keysAndValues...)
}

func (logger *Logger) Fatalf(template string, args ...interface{}) {
	logger.sugar.Fatalf(sprintf(template, args...))
}

func panicDetail(msg string, fields ...Field) {
	enc := zapcore.NewMapObjectEncoder()
	for _, field := range fields {
		field.AddTo(enc)
	}

	// 控制台输出
	fmt.Printf("%s: \n    %s: %s\n", Red("panic"), Red("msg"), msg)
	if _, file, line, ok := runtime.Caller(3); ok {
		fmt.Printf("    %s: %s:%d\n", Red("loc"), file, line)
	}
	for key, val := range enc.Fields {
		fmt.Printf("    %s: %s\n", Red(key), fmt.Sprintf("%+v", val))
	}

}

func (logger *Logger) With(fields ...Field) *Logger {
	deSugarLogger := logger.deSugar.With(fields...)
	return &Logger{
		deSugar: deSugarLogger,
		level:   logger.level,
		sugar:   deSugarLogger.Sugar(),
		config:  logger.config,
	}
}
