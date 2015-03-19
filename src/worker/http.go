package worker

import (
	"fmt"
	"net/http"
	"time"
)

type HttpResp struct {
	// Error   error
	Error   string
	Code    int
	Status  string
	TimeDur float64
}

func HttpRequest(url *string, method *string, verbose bool) (result HttpResp) {
	client := &http.Client{
		// CheckRedirect: redirectPolicyFunc,
		Timeout: time.Second * 3,
	}

	var response http.Response

	t1 := time.Now()

	switch *method {
	case "GET":
		resp, errhttp := client.Get(*url)
		if errhttp != nil {
			result.Error = errhttp.Error()
			return
		}
		defer resp.Body.Close()
		response = *resp
	case "HEAD":
		resp, errhttp := client.Head(*url)
		if errhttp != nil {
			result.Error = errhttp.Error()
			return
		}
		defer resp.Body.Close()
		response = *resp
	default:
		result.Error = fmt.Sprintf("method [%s] unsupport!", *method)
		return
	}

	t2 := time.Now()
	duration := t2.Sub(t1).Seconds()

	result.TimeDur = duration
	result.Code = response.StatusCode
	if verbose {
		result.Status = response.Status
	}

	return
}
