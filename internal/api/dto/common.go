package dto

type Response struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func Success(message string, data any) Response {
	return Response{OK: true, Message: message, Data: data}
}

func Error(message string) Response {
	return Response{OK: false, Error: message}
}

type MountResponse struct {
	ID            string `json:"id"`
	DiskPath      string `json:"diskPath"`
	PartitionName string `json:"partitionName"`
	PartitionType string `json:"partitionType"`
	Start         int32  `json:"start"`
	Size          int32  `json:"size"`
}
