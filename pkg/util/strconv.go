package util

import "strconv"

// StrToUint64 str -> uint64
func StrToUint64(str string) uint64 {
	i, _ := strconv.ParseUint(str, 10, 64)
	return i
}

// Uint64ToStr uint64 -> str
func Uint64ToStr(num uint64) string {
	return strconv.FormatUint(num, 10)
}

// Int64ToStr  int64 -> str
func Int64ToStr(num int64) string {
	return strconv.FormatInt(num, 10)
}

// StrToInt64 str -> int64
func StrToInt64(str string) int64 {
	i, _ := strconv.ParseInt(str, 10, 64)
	return i
}
