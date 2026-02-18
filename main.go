package main

import (
	"fmt"
	"os"
	"signet/cmd"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: signet <command> [options]")
		fmt.Fprintln(os.Stderr, "Commands: init, start, stop")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		cmd.RunInit(os.Args[2:])
	case "start":
		cmd.RunStart(os.Args[2:])
	case "stop":
		cmd.RunStop(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
