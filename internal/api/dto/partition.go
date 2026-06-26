package dto

type PartitionResponse struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Fit    string `json:"fit"`
	Start  int32  `json:"start"`
	Size   int32  `json:"size"`
	Status string `json:"status"`
}

type CreatePartitionRequest struct {
	Path string `json:"path"`
	Name string `json:"name"`
	Size int64  `json:"size"`
	Unit string `json:"unit"`
	Type string `json:"type"`
	Fit  string `json:"fit"`
}
