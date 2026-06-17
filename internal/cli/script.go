package cli

import (
	"bufio"
	"fmt"
	"os"
)

func (c *CLI) RunScript(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	c.dispatcher.SetScriptMode(true)
	defer c.dispatcher.SetScriptMode(false)

	fmt.Fprintf(c.out, "Ejecutando script: %s\n", path)

	scanner := bufio.NewScanner(file)
	lineNumber := 1
	for scanner.Scan() {
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
