package fs

import (
	"mia_p1_201800996/internal/structs"
)

type Actor struct {
	User string
	UID  int32
	GID  int32
}

func CanRead(inode structs.Inode, actor Actor) bool {
	return hasPermission(inode, actor, 4)
}

func CanWrite(inode structs.Inode, actor Actor) bool {
	return hasPermission(inode, actor, 2)
}

func IsRoot(actor Actor) bool {
	return actor.User == "root" || actor.UID == 1
}

func hasPermission(inode structs.Inode, actor Actor, mask int) bool {
	if IsRoot(actor) {
		return true
	}
	permIndex := 2
	if actor.UID == inode.IUid {
		permIndex = 0
	} else if actor.GID == inode.IGid {
		permIndex = 1
	}
	value := int(inode.IPerm[permIndex] - '0')
	return value&mask == mask
}
