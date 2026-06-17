package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"mia_p1_201800996/internal/commands"
	"mia_p1_201800996/internal/parser"
)

type CLI struct {
	dispatcher *commands.Dispatcher
	in         io.Reader
	out        io.Writer
}

func New(dispatcher *commands.Dispatcher, in io.Reader, out io.Writer) *CLI {
	return &CLI{
		dispatcher: dispatcher,
		in:         in,
		out:        out,
	}
}

func (c *CLI) RunREPL() error {
	fmt.Fprintln(c.out, "MIA Proyecto 1 CLI")
	fmt.Fprintln(c.out, "Escriba 'exit' para salir.")

	scanner := bufio.NewScanner(c.in)
	lineNumber := 1
	for {
		fmt.Fprint(c.out, "> ")
		if !scanner.Scan() {
			break
		}

		shouldExit, err := c.handleLine(scanner.Text(), lineNumber)
		if err != nil {
			fmt.Fprintln(c.out, err)
		}
		if shouldExit {
			return nil
		}
		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (c *CLI) handleLine(line string, lineNumber int) (bool, error) {
	cmd, skip, err := parser.ParseLine(strings.TrimRight(line, "\r\n"), lineNumber)
	if err != nil {
		return false, err
	}
	if skip {
		return false, nil
	}
	return c.dispatcher.Execute(cmd)
}
