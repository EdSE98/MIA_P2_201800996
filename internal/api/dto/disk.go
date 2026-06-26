package dto

type DiskResponse struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Size int64  `json:"size"`
}

type CreateDiskRequest struct {
	Size int64  `json:"size"`
	Unit string `json:"unit"`
	Fit  string `json:"fit"`
	Path string `json:"path"`
}

type DeleteDiskRequest struct {
	Path string `json:"path"`
}
