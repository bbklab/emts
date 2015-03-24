package inc

import (
	"regexp"
	"testing"
)

func Test_FGrepBool_1(t *testing.T) {
	if reg, err := regexp.Compile("^root:"); err == nil {
		file := "/etc/passwd"
		if !FGrepBool(file, reg) {
			t.Errorf("expect user root in %s", file)
			return
		}
	} else {
		t.Error(err)
		return
	}

	if reg, err := regexp.Compile("^ssh[ \t]*"); err == nil {
		file := "/etc/services"
		if !FGrepBool(file, reg) {
			t.Errorf("expect sshd service in %s", file)
			return
		}
	} else {
		t.Error(err)
		return
	}
}

func Test_FGrepLine_1(t *testing.T) {
	if reg, err := regexp.Compile("^root:"); err == nil {
		file := "/etc/passwd"
		line := FGrepLine(file, reg)
		if len(line) > 0 {
			t.Log("grep user root line:", string(line))
		} else {
			t.Errorf("expect user root in %s", file)
			return
		}
	} else {
		t.Error(err)
		return
	}

	if reg, err := regexp.Compile("^ssh[ \t]*"); err == nil {
		file := "/etc/services"
		line := FGrepLine(file, reg)
		if len(line) > 0 {
			t.Log("grep ssh service line:", string(line))
		} else {
			t.Errorf("expect sshd service in %s", file)
			return
		}
	} else {
		t.Error(err)
		return
	}
}
