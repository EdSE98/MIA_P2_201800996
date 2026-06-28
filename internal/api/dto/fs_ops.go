package dto

type EditFileRequest struct {
	Path      string `json:"path"`
	Contenido string `json:"contenido"`
}

type RenameEntryRequest struct {
	Path string `json:"path"`
	Name string `json:"name"`
}
