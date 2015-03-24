package inc

import (
	"testing"
	"time"
)

func Test_Caller_1(t *testing.T) {
	// single args
	today := Caller("date", "+%F_%T")
	if today != "" {
		t.Log(today)
	} else {
		t.Error("Single_Args: calller return empty")
		return
	}

	// multi args with space
	temp := "/etc/passwd /etc/services"
	ls := Caller("/bin/ls", temp)
	if ls != "" {
		t.Log(ls)
	} else {
		t.Error("Space_Args: caller return empty")
		return
	}
}

// test run by go
func Test_Caller_2(t *testing.T) {
	c := make(chan string)
	tmout := make(chan bool, 1)
	go func() {
		time.Sleep(3 * time.Second)
		tmout <- true
	}()

	go runCallerbyGo(c, "date", "+%F_%T")
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
}

func runCallerbyGo(c chan string, cmd, args string) {
	output := Caller(cmd, args)
	c <- output
}

func Test_Caller_3(t *testing.T) {
	for k, v := range Checker {
		t.Log(k, v)
	}
}
