package structs

import (
	"strings"
	"time"
)

const DateLayout = "2006-01-02 15:04:05"

func CopyString(dst []byte, value string) {
	for i := range dst {
		dst[i] = 0
	}
	copy(dst, []byte(value))
}

func FixedBytesToString(src []byte) string {
	return strings.TrimRight(string(src), "\x00")
}

func NewDateBytes(t time.Time) [20]byte {
	var result [20]byte
	CopyString(result[:], t.Format(DateLayout))
	return result
}

func NowDateBytes() [20]byte {
	return NewDateBytes(time.Now())
}

func SetName16(dst *[16]byte, value string) {
	CopyString(dst[:], value)
}

func SetName12(dst *[12]byte, value string) {
	CopyString(dst[:], value)
}

func SetID4(dst *[4]byte, value string) {
	CopyString(dst[:], value)
}

func SetPerm(dst *[3]byte, value string) {
	CopyString(dst[:], value)
}
