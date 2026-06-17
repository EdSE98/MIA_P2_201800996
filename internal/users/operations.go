package users

import "fmt"

const maxUserFieldLength = 10

func MakeGroup(content string, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("mkgrp requiere -name")
	}
	if len(name) > maxUserFieldLength {
		return "", fmt.Errorf("el nombre de grupo no puede exceder 10 caracteres")
	}

	records, err := ParseUsersRecords(content)
	if err != nil {
		return "", err
	}
	if activeGroupExists(records, name) {
		return "", fmt.Errorf("ya existe el grupo %q", name)
	}

	records = append(records, Record{
		ID:    nextGroupID(records),
		Type:  "G",
		Group: name,
	})
	return SerializeUsersFile(records), nil
}

func RemoveGroup(content string, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("rmgrp requiere -name")
	}
	if name == "root" {
		return "", fmt.Errorf("no se puede eliminar el grupo root")
	}

	records, err := ParseUsersRecords(content)
	if err != nil {
		return "", err
	}
	for i := range records {
		if records[i].Type == "G" && records[i].Active() && records[i].Group == name {
			records[i].ID = 0
			return SerializeUsersFile(records), nil
		}
	}
	return "", fmt.Errorf("el grupo no existe")
}

func MakeUser(content string, username string, password string, group string) (string, error) {
	if username == "" {
		return "", fmt.Errorf("mkusr requiere -user")
	}
	if password == "" {
		return "", fmt.Errorf("mkusr requiere -pass")
	}
	if group == "" {
		return "", fmt.Errorf("mkusr requiere -grp")
	}
	if len(username) > maxUserFieldLength || len(password) > maxUserFieldLength || len(group) > maxUserFieldLength {
		return "", fmt.Errorf("usuario, contraseña y grupo no pueden exceder 10 caracteres")
	}

	records, err := ParseUsersRecords(content)
	if err != nil {
		return "", err
	}
	if activeUserExists(records, username) {
		return "", fmt.Errorf("ya existe el usuario %q", username)
	}
	if !activeGroupExists(records, group) {
		return "", fmt.Errorf("el grupo no existe")
	}

	records = append(records, Record{
		ID:       nextUserID(records),
		Type:     "U",
		Group:    group,
		Username: username,
		Password: password,
	})
	return SerializeUsersFile(records), nil
}

func RemoveUser(content string, username string) (string, error) {
	if username == "" {
		return "", fmt.Errorf("rmusr requiere -user")
	}
	if username == "root" {
		return "", fmt.Errorf("no se puede eliminar el usuario root")
	}

	records, err := ParseUsersRecords(content)
	if err != nil {
		return "", err
	}
	for i := range records {
		if records[i].Type == "U" && records[i].Active() && records[i].Username == username {
			records[i].ID = 0
			return SerializeUsersFile(records), nil
		}
	}
	return "", fmt.Errorf("el usuario no existe")
}

func ChangeUserGroup(content string, username string, group string) (string, error) {
	if username == "" {
		return "", fmt.Errorf("chgrp requiere -user")
	}
	if group == "" {
		return "", fmt.Errorf("chgrp requiere -grp")
	}

	records, err := ParseUsersRecords(content)
	if err != nil {
		return "", err
	}
	if !activeGroupExists(records, group) {
		return "", fmt.Errorf("el grupo no existe")
	}
	for i := range records {
		if records[i].Type == "U" && records[i].Active() && records[i].Username == username {
			records[i].Group = group
			return SerializeUsersFile(records), nil
		}
	}
	return "", fmt.Errorf("el usuario no existe")
}

func activeGroupExists(records []Record, name string) bool {
	for _, record := range records {
		if record.Type == "G" && record.Active() && record.Group == name {
			return true
		}
	}
	return false
}

func activeUserExists(records []Record, username string) bool {
	for _, record := range records {
		if record.Type == "U" && record.Active() && record.Username == username {
			return true
		}
	}
	return false
}

func nextGroupID(records []Record) int32 {
	var max int32
	var count int32
	for _, record := range records {
		if record.Type != "G" {
			continue
		}
		count++
		if record.ID > max {
			max = record.ID
		}
	}
	if count > max {
		max = count
	}
	return max + 1
}

func nextUserID(records []Record) int32 {
	var max int32
	var count int32
	for _, record := range records {
		if record.Type != "U" {
			continue
		}
		count++
		if record.ID > max {
			max = record.ID
		}
	}
	if count > max {
		max = count
	}
	return max + 1
}
