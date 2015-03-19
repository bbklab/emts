package worker

import (
	"testing"
)

func Test_HttpRequest_1(t *testing.T) {
	array := [...][2]string{
		{"http://www.baidu.com", "GET"},
		{"https://www.alipay.com/", "HEAD"},
		// {"https://mail.163.com", "HEAD"},
		// {"http://www.sina.com", "HEAD"},
		// {"https://mail.126.com", "GET"},
		// {"https://mail.qq.com", "GET"},
	}

	for _, v := range array {
		url, method := v[0], v[1]
		result := HttpRequest(&url, &method, true)
		if result.Error != "" {
			t.Error(result.Error)
			continue
		}
		if result.Code != 200 {
			t.Errorf("%s(%s) response %s, expect 200", url, method, result.Code)
			continue
		}
		t.Logf("pass check: %s(%s) in %f response %d(%s)",
			url, method, result.TimeDur, result.Code, result.Status)
	}
}

func Test_HttpRequest_2(t *testing.T) {
	url, method := "https://passport.baidu.com/?q=login", "GET"
	result := HttpRequest(&url, &method, true)
	if result.Error != "" {
		t.Error(result.Error)
		return
	}
	if result.Code != 200 {
		t.Errorf("redirect url %s(%s) response %s, expect 200", url, method, result.Code)
		return
	}
	t.Logf("pass check: redirect url %s(%s) in %f response %d(%s)",
		url, method, result.TimeDur, result.Code, result.Status)
}
