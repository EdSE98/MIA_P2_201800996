package dto

type ReportRequest struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PathFileLs string `json:"pathFileLs"`
	Format     string `json:"format"`
}

type ReportResponse struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	ContentType string `json:"contentType"`
}
