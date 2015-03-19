package worker

import (
	"testing"
)

func Test_PingIP4_1(t *testing.T) {
	addr, num := "127.0.0.1", 5
	result := PingIP4(&addr, &num, true)

	if result.Error != "" {
		t.Error(result.Error)
		return
	}
	t.Log("pass check: ping localhost")

	if result.SendSum != num {
		t.Errorf("%d icmp sent, expect %d icmp sent", result.SendSum, num)
		return
	}
	t.Log("pass check: SendSum")

	if result.RecvSum != num {
		t.Errorf("%d response receive, expect %d response receive", result.RecvSum, num)
		return
	}
	t.Log("pass check: RecvSum")

	if result.AvgTime == -1 {
		t.Errorf("average == -1, receive 0 response")
		return
	}
	t.Log("pass check: AvgTime")
}
