package dto

type MountRequest struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

type UnmountRequest struct {
	ID string `json:"id"`
}

type MkfsRequest struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}
