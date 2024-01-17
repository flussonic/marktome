package main

import (
	"fmt"
	"foli2/md2json"
	"os"
)

type CommandFunction func([]string) error

var commands = map[string]CommandFunction{
	"md2json":        md2json.Commmand_md2json,
	"planarize":      md2json.Command_planarize,
	"superlinks":     md2json.Command_superlinks,
	"snippets":       md2json.Command_snippets,
	"macros":         md2json.Command_macros,
	"foliant2mkdocs": md2json.Commmand_foliant2mkdocs,
	"json2md":        md2json.Command_json2md,
}

func main() {
	if len(os.Args) >= 2 {
		cmd, ok := commands[os.Args[1]]
		if ok {
			err := cmd(os.Args[2:])
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(2)
			}
		} else {
			fmt.Printf("No such command: %s\n", os.Args[1])
			os.Exit(3)
		}
	}
}
