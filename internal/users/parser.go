package users

import (
	"fmt"
	"strconv"
	"strings"
)

type GroupRecord struct {
	ID     int32
	Name   string
	Active bool
}

type UserRecord struct {
	ID       int32
	Group    string
	Username string
	Password string
	Active   bool
}

func ParseUsersFile(content string) ([]GroupRecord, []UserRecord, error) {
	var groups []GroupRecord
	var users []UserRecord

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lineNumber := i + 1
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := splitFields(line)
		if len(fields) < 2 {
			return nil, nil, fmt.Errorf("linea %d de users.txt malformada", lineNumber)
		}

		id64, err := strconv.ParseInt(fields[0], 10, 32)
		if err != nil {
			return nil, nil, fmt.Errorf("linea %d de users.txt tiene ID invalido", lineNumber)
		}
		id := int32(id64)
		recordType := strings.ToUpper(fields[1])

		switch recordType {
		case "G":
			if len(fields) != 3 {
				return nil, nil, fmt.Errorf("linea %d de grupo malformada", lineNumber)
			}
			groups = append(groups, GroupRecord{
				ID:     id,
				Name:   fields[2],
				Active: id != 0,
			})
		case "U":
			if len(fields) != 5 {
				return nil, nil, fmt.Errorf("linea %d de usuario malformada", lineNumber)
			}
			users = append(users, UserRecord{
				ID:       id,
				Group:    fields[2],
				Username: fields[3],
				Password: fields[4],
				Active:   id != 0,
			})
		default:
			return nil, nil, fmt.Errorf("linea %d de users.txt tiene tipo desconocido %q", lineNumber, fields[1])
		}
	}

	return groups, users, nil
}

func FindActiveUser(content string, username string) (UserRecord, bool, error) {
	_, records, err := ParseUsersFile(content)
	if err != nil {
		return UserRecord{}, false, err
	}
	for _, record := range records {
		if record.Active && record.Username == username {
			return record, true, nil
		}
	}
	return UserRecord{}, false, nil
}

func FindActiveGroup(content string, groupName string) (GroupRecord, bool, error) {
	records, _, err := ParseUsersFile(content)
	if err != nil {
		return GroupRecord{}, false, err
	}
	for _, record := range records {
		if record.Active && record.Name == groupName {
			return record, true, nil
		}
	}
	return GroupRecord{}, false, nil
}

func GroupIDForUser(content string, user UserRecord) (int32, error) {
	group, ok, err := FindActiveGroup(content, user.Group)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, fmt.Errorf("el grupo %q del usuario %q no existe", user.Group, user.Username)
	}
	return group.ID, nil
}

func splitFields(line string) []string {
	rawFields := strings.Split(line, ",")
	fields := make([]string, len(rawFields))
	for i, field := range rawFields {
		fields[i] = strings.TrimSpace(field)
	}
	return fields
}
