package inc

import (
	"testing"
	"time"
)

func Test_Caller_1(t *testing.T) {
	today := Caller("date", "+%F_%T")
	if today != "" {
		t.Log(today)
	} else {
		t.Error(today)
		return
	}

	// test run by go
	c := make(chan string)
	tmout := make(chan bool, 1)
	go func() {
		time.Sleep(3 * time.Second)
		tmout <- true
	}()

	go runCallerbyGo(c, "date", "+%F_%T")
	//go runCallerbyGo(c, "sleep", "100")
	select {
	case output := <-c:
		t.Log("goroutine return:", output)
		if output == "" {
			t.Error("goroutine return empty")
			return
		}
	case <-tmout:
		t.Error("timeout")
		return
	}

	/*
		name := "/home/zgz/emts/src/c/dnsbl"
		args := "1.1.1.1 8.8.8.8 127.0.0.2"
		r := Caller(name, args)
		t.Log(r)
	*/
}

func runCallerbyGo(c chan string, cmd, args string) {
	output := Caller(cmd, args)
	c <- output
}

func Test_Caller_2(t *testing.T) {
	for k, v := range Checker {
		t.Log(k, v)
	}
}
