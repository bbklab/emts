package worker

import (
	"testing"
)

func Test_SmtpPing_1(t *testing.T) {
	array := [3]string{
		"126mx01.mxmail.netease.com:25",
		"mx1.qq.com:25",
		"163mx01.mxmail.netease.com:25",
	}

	for _, v := range array {
		result := SmtpPing(&v, true)
		if result.Error != "" {
			t.Error(result.Error)
			continue
		}
		if result.Code != 220 {
			t.Errorf("%s response %s, expect 220", v, result.Code)
			continue
		}
		t.Logf("pass check: %s in %f response %d(%s)",
			v, result.TimeDur, result.Code, result.Message)
	}
}

func Test_SmtpPing_2(t *testing.T) {
	addr := "0.0.0.0:22"
	result := SmtpPing(&addr, true)
	if result.Error == "" {
		t.Error(result.TimeDur, result.Code, result.Message)
		t.Error("should complain error here")
	}
	t.Logf("pass error check: %s return %s",
		addr, result.Error)
}
