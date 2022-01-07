package util

import (
	"encoding/json"
	"fmt"
	"math"
	"owl-engine/pkg/model/constParam"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func Float64ToString(input float64, prec int) string {
	return strconv.FormatFloat(input, 'f', prec, 64)
}

func Float64SlToStringSl(input []float64, prec int) (out []string) {
	for _, v := range input {
		tmp := Float64ToString(v, prec)
		out = append(out, tmp)
	}
	return out
}

func StringToInt(in string) int {
	out, _ := strconv.Atoi(in)
	return out
}

func IntSlToString(arrInt []int) string {
	str := strings.Replace(strings.Trim(fmt.Sprint(arrInt), "[]"), " ", ",", -1)
	return str
}

func StringToIntSl(str string) (arrInt []int) {
	strSl := strings.Split(str, constParam.SymbolComma)
	for _, v := range strSl {
		tmp, _ := strconv.Atoi(v)
		arrInt = append(arrInt, tmp)
	}
	return arrInt
}

func DateToString(dt time.Time) string {
	if dt.IsZero() {
		return constParam.NULL
	}
	return dt.Format(constParam.DateFormat)
}

func DateTimeToString(dt time.Time) string {
	if dt.IsZero() {
		return constParam.NULL
	}
	return dt.Format(constParam.DateTimeFormat)
}

func DatetimeToWholeMinutes(dt time.Time) string {
	if dt.IsZero() {
		return constParam.NULL
	}
	return dt.Format(constParam.DateTimeWholeMinFormat)
}

func StringToDate(dt string) (time.Time, error) {
	return time.ParseInLocation(constParam.DateFormat, dt, time.Local)
}

func StringToDateMonth(dt string) (time.Time, error) {
	return time.ParseInLocation(constParam.DateMonthFormat, dt, time.Local)
}

func StringToDateTime(dt string) (time.Time, error) {
	return time.ParseInLocation(constParam.DateTimeFormat, dt, time.Local)
}

func StringToDateTimeMin(dt string) (time.Time, error) {
	return time.ParseInLocation(constParam.DateTimeMinFormat, dt, time.Local)
}

func TimestampToDateStr(t int64) string {
	return time.Unix(t, 0).Format(constParam.DateFormat)
}

func TimestampToDateTimeStr(t int64) string {
	return time.Unix(t, 0).Format(constParam.DateTimeFormat)
}

func TimestampToDateTimeMinStr(t int64) string {
	return time.Unix(t, 0).Format(constParam.DateTimeMinFormat)
}

func FormatDtTimeStrToTimestamp(dt string) (int64, error) {
	date, err := time.ParseInLocation(constParam.DateTimeFormat, dt, time.Local)
	if err != nil {
		return 0, err
	}
	return date.Unix(), err
}

func AbsInt64(n int64) int64 {
	return int64(math.Abs(float64(n)))
}

func JsonToString(input interface{}) string {
	b, err := json.Marshal(input)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// 序列化数据
func Serialization(value interface{}) ([]byte, error) {
	if bytes, ok := value.([]byte); ok {
		return bytes, nil
	}

	switch v := reflect.ValueOf(value); v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return []byte(strconv.FormatInt(v.Int(), 10)), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return []byte(strconv.FormatUint(v.Uint(), 10)), nil
	case reflect.Map:
	}
	k, err := json.Marshal(value)
	return k, err
}

// 反序列化数据
func Deserialization(byt []byte, ptr interface{}) (err error) {
	if bytes, ok := ptr.(*[]byte); ok {
		*bytes = byt
		return
	}

	if v := reflect.ValueOf(ptr); v.Kind() == reflect.Ptr {
		switch p := v.Elem(); p.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			var i int64
			i, err = strconv.ParseInt(string(byt), 10, 64)
			if err != nil {
				return err
			} else {
				p.SetInt(i)
			}
			return

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			var i uint64
			i, err = strconv.ParseUint(string(byt), 10, 64)
			if err != nil {
				return err
			} else {
				p.SetUint(i)
			}
			return
		}
	}

	err = json.Unmarshal(byt, &ptr)
	return
}
