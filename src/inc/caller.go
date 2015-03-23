package inc

import (
	"os/exec"
)

func Caller(name string, args string) string {
	cmd := exec.Command(name, args)
	if output, err := cmd.Output(); err != nil {
		return ""
	} else {
		return string(output)
	}
}
