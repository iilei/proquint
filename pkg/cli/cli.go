package cli

import (
	"fmt"
	"os"
)

func Execute() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "help", "--help", "-h":
		printUsage()
	default:
		runProquint(os.Args[1:])
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "proquint encodes decimal and hex to 5-letter words")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  proquint <command> [options]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  encode")
	fmt.Fprintln(os.Stderr, "  decode")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Run 'proquint <command> --help' for more information on a command.")
}
