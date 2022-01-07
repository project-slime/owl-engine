package xlogs

import (
	"fmt"
	"math/rand"
	"strconv"
)

var _ = RandomColor()

func RandomColor() string {
	return fmt.Sprintf("#%s", strconv.FormatInt(int64(rand.Intn(16777216)), 16))
}

func Yellow(msg string, arg ...interface{}) string {
	return sprint(YellowColor, msg, arg...)
}

func Red(msg string, arg ...interface{}) string {
	return sprint(RedColor, msg, arg...)
}

func Blue(msg string, arg ...interface{}) string {
	return sprint(BlueColor, msg, arg...)
}

func Green(msg string, arg ...interface{}) string {
	return sprint(GreenColor, msg, arg...)
}

func Greenf(msg string, arg ...interface{}) string {
	return sprint(GreenColor, msg, arg...)
}

func sprint(colorValue int, msg string, arg ...interface{}) string {
	if arg != nil {
		return fmt.Sprintf("\x1b[%dm%s\x1b[0m %+v", colorValue, msg, arrToTransform(arg))
	} else {
		return fmt.Sprintf("\x1b[%dm%s\x1b[0m", colorValue, msg)
	}
}
