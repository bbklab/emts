package worker

import (
	"testing"
)

func Test_ImapPing_1(t *testing.T) {
	array := [3]string{
		"imap.126.com:143",
		"imap.163.com:143",
		"imap.qq.com:143",
	}

	for _, v := range array {
		result := ImapPing(&v, true)
		if result.Error != "" {
			t.Error(result.Error)
			continue
		}
		if result.Flag != "* OK" {
			t.Errorf("%s response %s, expect +OK", v, result.Flag)
			continue
		}
		t.Logf("pass check: %s in %f response %s(%s)",
			v, result.TimeDur, result.Flag, result.Message)
	}
}

func Test_ImapPing_2(t *testing.T) {
	addr := "0.0.0.0:22"
	result := ImapPing(&addr, true)
	if result.Error == "" {
		t.Error(result.TimeDur, result.Flag, result.Message)
		t.Error("should complain error here")
		return
	}
	t.Logf("pass error check: %s response %s",
		addr, result.Error)
}
