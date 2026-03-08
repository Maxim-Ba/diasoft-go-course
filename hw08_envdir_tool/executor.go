package main

import (
	"os"
	"os/exec"
	"strings"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) int {
	if len(cmd) == 0 {
		return 1
	}

	curEnv := os.Environ()
	envMap := make(map[string]string, len(curEnv))
	for _, kv := range curEnv {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	for name, ev := range env {
		if ev.NeedRemove {
			delete(envMap, name)
			// os.Unsetenv(name)
		} else {
			envMap[name] = ev.Value
		}
	}

	// Преобразуем map обратно в срез "key=value"
	envSlice := make([]string, 0, len(envMap))
	for k, v := range envMap {
		envSlice = append(envSlice, k+"="+v)
	}

	c := exec.Command(cmd[0], cmd[1:]...)
	c.Env = envSlice
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		return 1
	}
	return 0
}
