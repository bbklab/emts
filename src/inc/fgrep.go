package inc

import (
	"bufio"
	"os"
	"regexp"
)

func FGrepBool(file string, reg *regexp.Regexp) bool {
	if f, err := os.Open(file); err == nil {
		buf := bufio.NewReader(f)
		for {
			line, err := buf.ReadBytes('\n')
			if err != nil {
				return false
			}
			if reg.Match(line) {
				return true
			}
		}
	}
	return false
}

func FGrepLine(file string, reg *regexp.Regexp) []byte {
	if f, err := os.Open(file); err == nil {
		buf := bufio.NewReader(f)
		for {
			line, err := buf.ReadBytes('\n')
			if err != nil {
				return []byte("")
			}
			if reg.Match(line) {
				return line
			}
		}
	}
	return []byte("")
}
