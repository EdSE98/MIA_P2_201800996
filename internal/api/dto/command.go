package dto

type ExecuteCommandRequest struct {
	Command string `json:"command"`
}

type CommandExecutionResponse struct {
	Command string           `json:"command"`
	Output  string           `json:"output"`
	Session *SessionResponse `json:"session,omitempty"`
}
