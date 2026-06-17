package binio

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"reflect"
)

var byteOrder = binary.LittleEndian

func BinarySize(data any) (int64, error) {
	value, err := fixedSizeValue(data)
	if err != nil {
		return 0, err
	}

	size := binary.Size(value.Interface())
	if size < 0 {
		return 0, fmt.Errorf("dato de tamaño variable o no soportado: %T", data)
	}
	return int64(size), nil
}

func FileSize(file *os.File) (int64, error) {
	if file == nil {
		return 0, fmt.Errorf("archivo nil")
	}
	info, err := file.Stat()
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func EnsureRange(file *os.File, offset int64, size int64) error {
	if offset < 0 {
		return fmt.Errorf("offset invalido %d: no puede ser negativo", offset)
	}
	if size < 0 {
		return fmt.Errorf("tamaño invalido %d: no puede ser negativo", size)
	}

	fileSize, err := FileSize(file)
	if err != nil {
		return err
	}
	if offset > fileSize-size {
		return fmt.Errorf("rango fuera del archivo: offset=%d size=%d file_size=%d", offset, size, fileSize)
	}
	return nil
}

func ReadStructAt(file *os.File, offset int64, data any) error {
	size, err := BinarySize(data)
	if err != nil {
		return err
	}
	if err := EnsureRange(file, offset, size); err != nil {
		return err
	}

	buffer := make([]byte, size)
	if _, err := file.ReadAt(buffer, offset); err != nil {
		return err
	}

	return binary.Read(bytes.NewReader(buffer), byteOrder, data)
}

func WriteStructAt(file *os.File, offset int64, data any) error {
	size, err := BinarySize(data)
	if err != nil {
		return err
	}
	if err := EnsureRange(file, offset, size); err != nil {
		return err
	}

	var buffer bytes.Buffer
	if err := binary.Write(&buffer, byteOrder, data); err != nil {
		return err
	}
	if int64(buffer.Len()) != size {
		return fmt.Errorf("tamaño serializado inesperado: esperado=%d obtenido=%d", size, buffer.Len())
	}

	n, err := file.WriteAt(buffer.Bytes(), offset)
	if err != nil {
		return err
	}
	if n != buffer.Len() {
		return io.ErrShortWrite
	}
	return nil
}

func ReadBytesAt(file *os.File, offset int64, size int64) ([]byte, error) {
	if err := EnsureRange(file, offset, size); err != nil {
		return nil, err
	}

	buffer := make([]byte, size)
	if _, err := file.ReadAt(buffer, offset); err != nil {
		return nil, err
	}
	return buffer, nil
}

func WriteBytesAt(file *os.File, offset int64, data []byte) error {
	size := int64(len(data))
	if err := EnsureRange(file, offset, size); err != nil {
		return err
	}

	n, err := file.WriteAt(data, offset)
	if err != nil {
		return err
	}
	if n != len(data) {
		return io.ErrShortWrite
	}
	return nil
}

func fixedSizeValue(data any) (reflect.Value, error) {
	if data == nil {
		return reflect.Value{}, fmt.Errorf("dato nil")
	}

	value := reflect.ValueOf(data)
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return reflect.Value{}, fmt.Errorf("puntero nil: %T", data)
		}
		value = value.Elem()
	}

	return value, nil
}
