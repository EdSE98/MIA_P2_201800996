package explorer

type Item struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"`
	Size        int32  `json:"size"`
	Inode       int32  `json:"inode"`
	Permissions string `json:"permissions"`
	Owner       string `json:"owner"`
	Group       string `json:"group"`
}

type DirectoryListing struct {
	ID    string `json:"id"`
	Path  string `json:"path"`
	Items []Item `json:"items"`
}

type FileContent struct {
	ID      string `json:"id"`
	Path    string `json:"path"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Size    int32  `json:"size"`
}

type Metadata struct {
	ID          string `json:"id"`
	Path        string `json:"path"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Inode       int32  `json:"inode"`
	Size        int32  `json:"size"`
	Permissions string `json:"permissions"`
	Owner       string `json:"owner"`
	Group       string `json:"group"`
}
