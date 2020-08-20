package bbio

import (
	"fmt"
	"strconv"
	"strings"
)

type cast struct{}

// Cast instance
var Cast cast

func stringToInt64(str string) int64 {
	var ret int64
	cleaned := strings.Replace(str, "0x", "", -1)
	if str == cleaned {
		i, err := strconv.ParseInt(str, 10, 64)
		if err == nil {
			ret = i
		}
	} else {
		i, err := strconv.ParseInt(str, 16, 64)
		if err == nil {
			ret = i
		}
	}
	return ret
}

// Int32 implements for Cast
func (cast) Int32(val interface{}) int32 {
	var ret int32

	switch t := val.(type) {
	case int:
		ret = int32(t)
	case int8:
		ret = int32(t)
	case int16:
		ret = int32(t)
	case int32:
		ret = t
	case int64:
		ret = int32(t)
	case bool:
		if t == true {
			ret = int32(1)
		} else {
			ret = int32(0)
		}
	case float32:
		ret = int32(t)
	case float64:
		ret = int32(t)
	case uint8:
		ret = int32(t)
	case uint16:
		ret = int32(t)
	case uint32:
		ret = int32(t)
	case uint64:
		ret = int32(t)
	case string:
		tmp := stringToInt64(val.(string))
		ret = int32(tmp)
	default:
		str := fmt.Sprintf("%s", val)
		tmp := stringToInt64(str)
		ret = int32(tmp)
	}

	return ret
}

// Int64 implements for Cast
func (cast) Int64(val interface{}) int64 {
	var ret int64

	switch t := val.(type) {
	case int:
		ret = int64(t)
	case int8:
		ret = int64(t)
	case int16:
		ret = int64(t)
	case int32:
		ret = int64(t)
	case int64:
		ret = t
	case bool:
		if t == true {
			ret = int64(1)
		} else {
			ret = int64(0)
		}
	case float32:
		ret = int64(t)
	case float64:
		ret = int64(t)
	case uint8:
		ret = int64(t)
	case uint16:
		ret = int64(t)
	case uint32:
		ret = int64(t)
	case uint64:
		ret = int64(t)
	case string:
		ret = stringToInt64(val.(string))
	default:
		str := fmt.Sprintf("%s", val)
		ret = stringToInt64(str)
	}

	return ret
}
