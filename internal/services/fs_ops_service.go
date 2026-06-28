package services

import (
	"mia_p1_201800996/internal/api/dto"
	"mia_p1_201800996/internal/fs"
	"mia_p1_201800996/internal/session"
)

func EditFile(req dto.EditFileRequest) error {
	active, actor, err := activeFSContext()
	if err != nil {
		return err
	}
	return fs.Edit(active.DiskPath, int64(active.PartitionStart), req.Path, req.Contenido, actor)
}

func RenameEntry(req dto.RenameEntryRequest) error {
	active, actor, err := activeFSContext()
	if err != nil {
		return err
	}
	return fs.Rename(active.DiskPath, int64(active.PartitionStart), req.Path, req.Name, actor)
}

func RemoveEntry(req dto.RemoveEntryRequest) error {
	active, actor, err := activeFSContext()
	if err != nil {
		return err
	}
	return fs.Remove(active.DiskPath, int64(active.PartitionStart), req.Path, actor)
}

func CopyEntry(req dto.TransferEntryRequest) (dto.CopyEntryResponse, error) {
	active, actor, err := activeFSContext()
	if err != nil {
		return dto.CopyEntryResponse{}, err
	}
	warnings, err := fs.Copy(active.DiskPath, int64(active.PartitionStart), req.Path, req.Destino, actor)
	return dto.CopyEntryResponse{Warnings: warnings}, err
}

func MoveEntry(req dto.TransferEntryRequest) error {
	active, actor, err := activeFSContext()
	if err != nil {
		return err
	}
	return fs.Move(active.DiskPath, int64(active.PartitionStart), req.Path, req.Destino, actor)
}

func activeFSContext() (session.Session, fs.Actor, error) {
	active, err := session.RequireActive()
	if err != nil {
		return session.Session{}, fs.Actor{}, err
	}
	return active, fs.Actor{User: active.User, UID: active.UID, GID: active.GID}, nil
}
