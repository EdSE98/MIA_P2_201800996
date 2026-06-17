package parser

import (
	"fmt"
	"strings"
	"unicode"

	"mia_p1_201800996/internal/commands"
)

var commandAliases = map[string]string{
	"mkuser": "mkusr",
}

var paramAliases = map[string]string{
	"s":     "size",
	"usr":   "user",
	"grupo": "grp",
}

var fitAliases = map[string]string{
	"firstfit": "FF",
	"bestfit":  "BF",
	"worstfit": "WF",
}

var reportAliases = map[string]string{
	"bm_bloc": "bm_block",
}

func ParseLine(input string, lineNumber int) (commands.Command, bool, error) {
	raw := strings.TrimRight(input, "\r\n")
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return commands.Command{}, true, nil
	}

	if strings.HasPrefix(trimmed, "#") {
		return commands.Command{
			Name:   commands.CommentCommand,
			Params: map[string]string{},
			Flags:  map[string]bool{},
			Raw:    raw,
			Line:   lineNumber,
		}, false, nil
	}

	tokens, err := tokenize(trimmed)
	if err != nil {
		return commands.Command{}, false, fmt.Errorf("linea %d: %w", lineNumber, err)
	}
	if len(tokens) == 0 {
		return commands.Command{}, true, nil
	}

	name := normalizeCommand(tokens[0])
	if strings.HasPrefix(name, "-") {
		return commands.Command{}, false, fmt.Errorf("linea %d: falta el nombre del comando", lineNumber)
	}

	cmd := commands.Command{
		Name:   name,
		Params: map[string]string{},
		Flags:  map[string]bool{},
		Raw:    raw,
		Line:   lineNumber,
	}

	for _, token := range tokens[1:] {
		key, value, hasValue, err := parseParamToken(token)
		if err != nil {
			return commands.Command{}, false, fmt.Errorf("linea %d: %w", lineNumber, err)
		}
		key = normalizeParam(key)
		if hasValue {
			cmd.Params[key] = normalizeValue(cmd.Name, key, value)
			continue
		}
		cmd.Flags[key] = true
	}

	return cmd, false, nil
}

func tokenize(input string) ([]string, error) {
	var tokens []string
	var current strings.Builder
	inQuote := false
	tokenStarted := false

	for _, ch := range input {
		switch {
		case ch == '"':
			inQuote = !inQuote
			tokenStarted = true
		case ch == '#' && !inQuote:
			if current.Len() > 0 || tokenStarted {
				tokens = append(tokens, current.String())
			}
			return tokens, nil
		case unicode.IsSpace(ch) && !inQuote:
			if current.Len() > 0 || tokenStarted {
				tokens = append(tokens, current.String())
				current.Reset()
				tokenStarted = false
			}
		default:
			current.WriteRune(ch)
			tokenStarted = true
		}
	}

	if inQuote {
		return nil, fmt.Errorf("comillas dobles sin cerrar")
	}
	if current.Len() > 0 || tokenStarted {
		tokens = append(tokens, current.String())
	}
	return tokens, nil
}

func parseParamToken(token string) (string, string, bool, error) {
	if !strings.HasPrefix(token, "-") {
		return "", "", false, fmt.Errorf("token invalido %q: los parametros deben iniciar con '-'", token)
	}

	body := strings.TrimLeft(token, "-")
	if body == "" {
		return "", "", false, fmt.Errorf("token invalido %q", token)
	}

	index := strings.Index(body, "=")
	if index == -1 {
		if !isValidName(body) {
			return "", "", false, fmt.Errorf("flag invalido %q", token)
		}
		return body, "", false, nil
	}

	key := body[:index]
	value := body[index+1:]

	// Tolerancia para entradas del script como -type-=full o -id-=XXXX.
	key = strings.TrimSuffix(key, "-")
	// Tolerancia para entradas como -grp==root.
	value = strings.TrimLeft(value, "=")

	if key == "" || !isValidName(key) {
		return "", "", false, fmt.Errorf("parametro invalido %q", token)
	}

	return key, value, true, nil
}

func normalizeCommand(name string) string {
	normalized := strings.ToLower(name)
	if alias, ok := commandAliases[normalized]; ok {
		return alias
	}
	return normalized
}

func normalizeParam(name string) string {
	normalized := strings.ToLower(name)
	if alias, ok := paramAliases[normalized]; ok {
		return alias
	}
	return normalized
}

func normalizeValue(command, key, value string) string {
	if key == "fit" {
		if alias, ok := fitAliases[strings.ToLower(value)]; ok {
			return alias
		}
	}
	if command == "rep" && key == "name" {
		if alias, ok := reportAliases[strings.ToLower(value)]; ok {
			return alias
		}
	}
	return value
}

func isValidName(value string) bool {
	for _, ch := range value {
		if ch == '_' || unicode.IsLetter(ch) || unicode.IsDigit(ch) {
			continue
		}
		return false
	}
	return true
}
