package main

import (
	"fmt"
	"os"
)

const exitCode = 111

func main() {
	self := os.Args[0]
	args := os.Args[1:]
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, `Not enough arguments
Usage: %s path_to_env_dir path_to_util [util_argument, ...]
`, self)
		os.Exit(exitCode)
	}

	environment, err := ReadDir(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitCode)
	}

	returnCode := RunCmd(args[1:], environment)

	os.Exit(returnCode)
}
