package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args

	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage is: go-envdir /path/to/env/dir <command> [args...]")
		os.Exit(1)
	}
	dir := args[1]
	cmd := args[2:]
	if len(cmd) == 0 {
		fmt.Fprintln(os.Stderr, "no command specified")
		os.Exit(1)
	}
	envs, err := ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading dir: %v\n", err)
		os.Exit(1)
	}

	code := RunCmd(cmd, envs)
	os.Exit(code)
}
