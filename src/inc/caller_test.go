package inc

import (
	"testing"
)

func Test_Caller_1(t *testing.T) {
	today := Caller("date", "+%F_%T")
	if today != "" {
		t.Log(today)
	} else {
		t.Error(today)
		return
	}
}

func Test_Caller_2(t *testing.T) {
	for k, v := range Checker {
		t.Log(k, v)
	}
}
