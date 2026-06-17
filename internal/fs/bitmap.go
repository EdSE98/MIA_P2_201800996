package fs

import (
	"fmt"
	"os"

	"mia_p1_201800996/internal/binio"
)

func ReadBitmap(file *os.File, start int64, count int32) ([]byte, error) {
	if count < 0 {
		return nil, fmt.Errorf("cantidad de bitmap invalida")
	}
	return binio.ReadBytesAt(file, start, int64(count))
}

func WriteBitmap(file *os.File, start int64, bitmap []byte) error {
	return binio.WriteBytesAt(file, start, bitmap)
}

func MarkBitmapUsed(bitmap []byte, index int32) error {
	if index < 0 || int(index) >= len(bitmap) {
		return fmt.Errorf("indice de bitmap fuera de rango: %d", index)
	}
	bitmap[index] = 1
	return nil
}

func FindFirstFree(bitmap []byte) (int32, bool) {
	for i, value := range bitmap {
		if value == 0 {
			return int32(i), true
		}
	}
	return -1, false
}
