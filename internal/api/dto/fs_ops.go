package dto

type EditFileRequest struct {
	Path      string `json:"path"`
	Contenido string `json:"contenido"`
}

type RenameEntryRequest struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

type RemoveEntryRequest struct {
	Path string `json:"path"`
}

type TransferEntryRequest struct {
	Path    string `json:"path"`
	Destino string `json:"destino"`
}

type CopyEntryResponse struct {
	Warnings []string `json:"warnings,omitempty"`
}
