package inc

import (
	"os"
	"os/exec"
	"path/filepath"
)

func GetAppRealPath() string {
	if path, err := exec.LookPath(os.Args[0]); err == nil {
		if rpath, err := filepath.Abs(path); err == nil {
			return rpath
		}
	}
	return ""
}
