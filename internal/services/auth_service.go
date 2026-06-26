package services

import (
	"mia_p1_201800996/internal/api/dto"
	"mia_p1_201800996/internal/session"
)

func Login(id string, user string, password string) (dto.SessionResponse, error) {
	active, err := session.Login(user, password, id)
	if err != nil {
		return dto.SessionResponse{}, err
	}
	return sessionResponse(active), nil
}

func Logout() error {
	return session.Logout()
}

func sessionResponse(active session.Session) dto.SessionResponse {
	return dto.SessionResponse{
		Active:        active.Active,
		MountedID:     active.MountedID,
		User:          active.User,
		UID:           active.UID,
		Group:         active.Group,
		GID:           active.GID,
		DiskPath:      active.DiskPath,
		PartitionName: active.PartitionName,
	}
}
