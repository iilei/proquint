package main

import (
	"fmt"
	"os"

	"github.com/iilei/proquint/pkg/cli"
)

var version = "dev"

func main() {
	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "--version", "-v", "version":
			fmt.Println(version)
			return
		}
	}

	cli.Execute()
}
