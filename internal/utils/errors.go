package utils

import "fmt"

func LineError(line int, format string, args ...any) error {
	return fmt.Errorf("linea %d: %s", line, fmt.Sprintf(format, args...))
}
