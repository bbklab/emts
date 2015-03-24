package inc

import (
	"os/exec"
)

func Caller(name string, args string) string {
	//cmd := exec.Command(name, args)	// this lead to error if args contains space
	cmd := exec.Command("/bin/sh", "-c", name+" "+args) // this is all right
	if output, err := cmd.Output(); err != nil {
		return ""
	} else {
		return string(output)
	}
}
