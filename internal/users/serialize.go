package users

import (
	"fmt"
	"strconv"
	"strings"
)

type Record struct {
	ID       int32
	Type     string
	Group    string
	Username string
	Password string
}

func ParseUsersRecords(content string) ([]Record, error) {
	var records []Record
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lineNumber := i + 1
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := splitFields(line)
		if len(fields) < 2 {
			return nil, fmt.Errorf("linea %d de users.txt malformada", lineNumber)
		}
		id64, err := strconv.ParseInt(fields[0], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("linea %d de users.txt tiene ID invalido", lineNumber)
		}

		record := Record{
			ID:   int32(id64),
			Type: strings.ToUpper(fields[1]),
		}
		switch record.Type {
		case "G":
			if len(fields) != 3 {
				return nil, fmt.Errorf("linea %d de grupo malformada", lineNumber)
			}
			record.Group = fields[2]
		case "U":
			if len(fields) != 5 {
				return nil, fmt.Errorf("linea %d de usuario malformada", lineNumber)
			}
			record.Group = fields[2]
			record.Username = fields[3]
			record.Password = fields[4]
		default:
			return nil, fmt.Errorf("linea %d de users.txt tiene tipo desconocido %q", lineNumber, fields[1])
		}
		records = append(records, record)
	}
	return records, nil
}

func SerializeUsersFile(records []Record) string {
	var b strings.Builder
	for _, record := range records {
		switch record.Type {
		case "G":
			fmt.Fprintf(&b, "%d,G,%s\n", record.ID, record.Group)
		case "U":
			fmt.Fprintf(&b, "%d,U,%s,%s,%s\n", record.ID, record.Group, record.Username, record.Password)
		}
	}
	return b.String()
}

func (r Record) Active() bool {
	return r.ID != 0
}
