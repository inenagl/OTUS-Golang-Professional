package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

func processEnv(env Environment) ([]string, error) {
	for envVar, envValue := range env {
		if envValue.NeedRemove {
			if err := os.Unsetenv(envVar); err != nil {
				return []string{}, err
			}
		} else {
			if err := os.Setenv(envVar, envValue.Value); err != nil {
				return []string{}, err
			}
		}
	}
	return os.Environ(), nil
}

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	if len(cmd) == 0 {
		fmt.Fprintln(os.Stderr, "RunCmd: no command name is given")
		return exitCode
	}

	command := exec.Command(cmd[0], cmd[1:]...) //nolint: gosec
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	environment, err := processEnv(env)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return exitCode
	}
	command.Env = environment

	if err = command.Start(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return exitCode
	}

	if err = command.Wait(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return exitError.ExitCode()
		}
		return exitCode
	}

	return 0
}
