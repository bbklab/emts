package worker

import (
	"testing"
)

func Test_PopPing_1(t *testing.T) {
	array := [...]string{
		"pop.126.com:110",
		"pop.163.com:110",
		"pop.qq.com:110",
		"mail.yili.com:110",
	}

	for _, v := range array {
		result := PopPing(&v, true)
		if result.Error != "" {
			t.Error(result.Error)
			continue
		}
		if result.Flag != "+OK" {
			t.Errorf("%s response %s, expect +OK", v, result.Flag)
			continue
		}
		t.Logf("pass check: %s in %f response %s(%s)",
			v, result.TimeDur, result.Flag, result.Message)
	}
}

func Test_PopPing_2(t *testing.T) {
	addr := "0.0.0.0:22"
	result := PopPing(&addr, true)
	if result.Error == "" {
		t.Error(result.TimeDur, result.Flag, result.Message)
		t.Error("should complain error here")
	}
	t.Logf("pass error check: %s return error %s",
		addr, result.Error)
}
