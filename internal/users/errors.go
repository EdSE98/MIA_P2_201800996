package users

import "errors"

var (
	ErrUserNotFound  = errors.New("el usuario no existe")
	ErrBadPassword   = errors.New("autenticacion fallida")
	ErrGroupNotFound = errors.New("el grupo del usuario no existe")
)
