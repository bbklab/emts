package worker

import (
	"testing"
)

func Test_MemcachePing_1(t *testing.T) {
	array := [...]string{
		"127.0.0.1:11211",
		"127.0.0.1:11231",
		"127.0.0.1:11251",
	}

	for _, v := range array {
		result := MemcachePing(&v, true)
		if result.Error != "" {
			t.Error(result.Error)
			continue
		}
		t.Logf("pass check: %s in %f response %v",
			v, result.TimeDur, result.Stat)
	}
}
