package xlogs

import (
	"io"

	"owl-engine/pkg/xlogs/rotate"
)

func newRotate(config *LoggerConfig) io.Writer {
	rotateLog := rotate.NewLogger()
	rotateLog.Filename = config.Filename()
	rotateLog.MaxSize = config.MaxSize
	rotateLog.MaxAge = config.MaxAge
	rotateLog.MaxBackups = config.MaxBackup
	rotateLog.Interval = config.Interval
	rotateLog.LocalTime = true
	rotateLog.Compress = config.Compress

	return rotateLog
}
