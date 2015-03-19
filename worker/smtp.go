package worker

import (
	"bufio"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type SmtpResp struct {
	// Error   error
	Error   string
	Code    int
	Message string
	TimeDur float64
}

func isLikeSmtpResp(s *string) bool {
	news := strings.TrimRight(*s, "\r\n") // !!! trim first
	smtpregex := regexp.MustCompile(`^[0-9]{3}[ -].+$`)
	return smtpregex.MatchString(news)
}

func parseCodeLine(s string) (code int, message string, err error) {
	if len(s) < 4 || (s[3] != ' ' && s[3] != '-') { // !!! must be ' ' not " "
		err = fmt.Errorf(s)
		return
	}

	code, errconv := strconv.Atoi(s[0:3])
	if errconv != nil {
		err = fmt.Errorf("convert code as int error: [%s]", errconv.Error())
		return
	}
	if code < 100 {
		err = fmt.Errorf("code [%d] deformity", code)
		return
	}

	message = s[4:]
	return
}

func SmtpPing(addr *string, verbose bool) (result SmtpResp) {

	t1 := time.Now()

	conn, errdial := net.Dial("tcp", *addr)
	if errdial != nil {
		result.Error = errdial.Error()
		return
	}
	defer func() {
		conn.Close()
	}()

	welcome, errread := bufio.NewReader(conn).ReadString('\n') // !!! must be '\n' as ReadString() want byte
	if errread != nil {
		result.Error = fmt.Sprintf("read line on server response failed: [%s]", errread.Error())
		return
	}

	welcome = strings.TrimSpace(welcome)

	if !(isLikeSmtpResp(&welcome)) {
		result.Error = fmt.Sprintf("seems unlike smtp response [%s]", welcome)
		return
	}

	code, message, errparse := parseCodeLine(welcome)
	if errparse != nil {
		result.Error = fmt.Sprintf("parse welcome response failed: [%s]", errparse.Error())
		return
	}
	if code != 220 {
		result.Error = fmt.Sprintf(welcome)
		return
	}

	t2 := time.Now()

	duration := t2.Sub(t1).Seconds()
	result.TimeDur = duration

	result.Code = code
	if verbose {
		result.Message = strings.TrimRight(message, "\r\n")
	}

	return
}
