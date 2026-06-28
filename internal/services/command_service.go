package services

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	"mia_p1_201800996/internal/api/dto"
	"mia_p1_201800996/internal/commands"
	"mia_p1_201800996/internal/parser"
	"mia_p1_201800996/internal/session"
)

var commandExecutionMu sync.Mutex

func ExecuteCommand(input string) (dto.CommandExecutionResponse, error) {
	commandExecutionMu.Lock()
	defer commandExecutionMu.Unlock()

	if strings.TrimSpace(input) == "" {
		return dto.CommandExecutionResponse{}, fmt.Errorf("comando vacio")
	}
	if strings.ContainsAny(input, "\r\n") {
		return dto.CommandExecutionResponse{}, fmt.Errorf("solo se permite una linea de comando")
	}

	parsed, skip, err := parser.ParseLine(input, 1)
	if err != nil {
		return dto.CommandExecutionResponse{}, err
	}
	if skip || parsed.IsComment() {
		return dto.CommandExecutionResponse{}, fmt.Errorf("no hay un comando ejecutable")
	}
	if parsed.Name == "pause" || parsed.Name == "exit" {
		return dto.CommandExecutionResponse{}, fmt.Errorf("el comando %s no esta permitido en la consola web", parsed.Name)
	}

	var output bytes.Buffer
	dispatcher := commands.NewDispatcher(strings.NewReader(""), &output)
	dispatcher.SetScriptMode(true)
	shouldExit, err := dispatcher.Execute(parsed)
	if err != nil {
		return dto.CommandExecutionResponse{}, err
	}
	if shouldExit {
		return dto.CommandExecutionResponse{}, fmt.Errorf("el comando no puede cerrar el servidor")
	}

	result := dto.CommandExecutionResponse{
		Command: input,
		Output:  strings.TrimRight(output.String(), "\r\n"),
	}
	if active, ok := session.Current(); ok {
		current := sessionResponse(active)
		result.Session = &current
	}
	return result, nil
}
