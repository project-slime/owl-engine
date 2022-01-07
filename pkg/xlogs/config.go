package xlogs

var Log *Logger

// Flush ...
func Flush() {
	Log.Flush()
}

// Info ...
func Info(msg string, fields ...Field) {
	Log.Info(msg, fields...)
}

// Debug ...
func Debug(msg string, fields ...Field) {
	Log.Debug(msg, fields...)
}

// Warn ...
func Warn(msg string, fields ...Field) {
	Log.Warn(msg, fields...)
}

// Error ...
func Error(msg string, fields ...Field) {
	Log.Error(msg, fields...)
}

// Panic ...
func Panic(msg string, fields ...Field) {
	Log.Panic(msg, fields...)
}

// DPanic ...
func DPanic(msg string, fields ...Field) {
	Log.DPanic(msg, fields...)
}

// Fatal ...
func Fatal(msg string, fields ...Field) {
	Log.Fatal(msg, fields...)
}

// Debugw ...
func Debugw(msg string, keysAndValues ...interface{}) {
	Log.Debugw(msg, keysAndValues...)
}

// Infow ...
func Infow(msg string, keysAndValues ...interface{}) {
	Log.Infow(msg, keysAndValues...)
}

// Warnw ...
func Warnw(msg string, keysAndValues ...interface{}) {
	Log.Warnw(msg, keysAndValues...)
}

// Errorw ...
func Errorw(msg string, keysAndValues ...interface{}) {
	Log.Errorw(msg, keysAndValues...)
}

// Panicw ...
func Panicw(msg string, keysAndValues ...interface{}) {
	Log.Panicw(msg, keysAndValues...)
}

// DPanicw ...
func DPanicw(msg string, keysAndValues ...interface{}) {
	Log.DPanicw(msg, keysAndValues...)
}

// Fatalw ...
func Fatalw(msg string, keysAndValues ...interface{}) {
	Log.Fatalw(msg, keysAndValues...)
}

// Debugf ...
func Debugf(msg string, args ...interface{}) {
	Log.Debugf(msg, args...)
}

// Infof ...
func Infof(msg string, args ...interface{}) {
	Log.Infof(msg, args...)
}

// Warnf ...
func Warnf(msg string, args ...interface{}) {
	Log.Warnf(msg, args...)
}

// Errorf ...
func Errorf(msg string, args ...interface{}) {
	Log.Errorf(msg, args...)
}

// Panicf ...
func Panicf(msg string, args ...interface{}) {
	Log.Panicf(msg, args...)
}

// DPanicf ...
func DPanicf(msg string, args ...interface{}) {
	Log.DPanicf(msg, args...)
}

// Fatalf ...
func Fatalf(msg string, args ...interface{}) {
	Log.Fatalf(msg, args...)
}

// Log ...
func (fn Func) Log(msg string, fields ...Field) {
	fn(msg, fields...)
}

// With ...
func With(fields ...Field) *Logger {
	return Log.With(fields...)
}
