package reports

import "strings"

func BuildBitmapText(bitmap []byte) string {
	var b strings.Builder
	for i, value := range bitmap {
		if i > 0 {
			if i%20 == 0 {
				b.WriteByte('\n')
			} else {
				b.WriteByte(' ')
			}
		}
		if value == 1 {
			b.WriteByte('1')
		} else {
			b.WriteByte('0')
		}
	}
	b.WriteByte('\n')
	return b.String()
}
