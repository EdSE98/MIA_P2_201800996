package main

import (
	"flag"
	"fmt"
	"os"

	"mia_p1_201800996/internal/cli"
	"mia_p1_201800996/internal/commands"
)

func main() {
	scriptPath := flag.String("script", "", "ruta del script .smia o .bat a ejecutar")
	flag.Parse()

	dispatcher := commands.NewDispatcher(os.Stdin, os.Stdout)
	runner := cli.New(dispatcher, os.Stdin, os.Stdout)

	var err error
	if *scriptPath != "" {
		err = runner.RunScript(*scriptPath)
	} else {
		err = runner.RunREPL()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
