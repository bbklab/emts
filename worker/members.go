package worker

import (
	"fmt"
)

var (
	Members = []string{
		"ping",
		"dns",
		"port",
		"http",
		"imap",
		"smtp",
		"pop",
	}
)

func ShowMembers() {
	fmt.Println(Members)
}
