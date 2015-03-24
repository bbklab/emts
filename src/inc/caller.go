package inc

import (
	"os/exec"
)

func Caller(command string, args []string) string {
	// cmd := exec.Command(command, args) // this lead to error if args contains space
	cmd := exec.Command(command, args...) // this is all right
	if output, err := cmd.Output(); err != nil {
		return ""
	} else {
		return string(output)
	}
}
