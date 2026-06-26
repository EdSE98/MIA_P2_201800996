package dto

type LoginRequest struct {
	ID       string `json:"id"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type SessionResponse struct {
	Active        bool   `json:"active"`
	MountedID     string `json:"mountedId"`
	User          string `json:"user"`
	UID           int32  `json:"uid"`
	Group         string `json:"group"`
	GID           int32  `json:"gid"`
	DiskPath      string `json:"diskPath"`
	PartitionName string `json:"partitionName"`
}
