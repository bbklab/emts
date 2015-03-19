package worker

import (
	"testing"
)

func Test_Resolve_1(t *testing.T) {
	array := [6][2]string{
		{"126.com", "ns"},
		{"126.com", "mx"},
		{"126.com", "txt"},
		{"8.8.8.8", "ptr"},
		{"mail.126.com", "a"},
		{"mail.126.com", "cname"},
	}
	for _, v := range array {
		addr, rtype := v[0], v[1]
		result := Resolve(&addr, &rtype, true)
		if result.Error != "" {
			t.Error(result.Error)
			continue
		}
		if !(len(result.Body) > 0) {
			t.Errorf("resolve %s (%s) result empty", addr, rtype)
			continue
		}
		t.Logf("pass resolve in %f, %s (%s) = %s ",
			result.TimeDur, addr, rtype, result.Body)
	}
}
