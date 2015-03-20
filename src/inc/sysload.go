package inc

import (
	"bufio"
	"os"
	"strconv"
)

func GetSysLoadavg() int {
	file, err := os.Open("/proc/loadavg")
	if err != nil {
		return 0
	}
	defer file.Close()

	buf := bufio.NewReader(file)
	if load, err := buf.ReadBytes(' '); err != nil {
		return 0
	} else {
		if result, err := strconv.Atoi(string(load)); err != nil {
			return 0
		} else {
			return result
		}
	}
}
