package inc

import (
	"testing"
)

func Test_GetSysLoadavg_1(t *testing.T) {
	load := GetSysLoadavg()
	if load >= 0 {
		t.Log("loadavg: ", load)
	} else {
		t.Error("loadavg: ", load)
	}
}
