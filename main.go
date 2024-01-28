package main

import (
	"fmt"
	"marktome/md2json"
	"os"
)

func main() {
	if len(os.Args) >= 2 {
		cmd, ok := md2json.Commands[os.Args[1]]
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
