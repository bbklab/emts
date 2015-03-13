package inc

import (
	"os"
	"testing"
)

func Test_SearchFile_1(t *testing.T) {
	dir := "./.testdir"
	err := os.Mkdir(dir, 0777)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		os.RemoveAll(dir)
	}()

	dirs := []string{"dir1", "dir2", "dir3"}
	files := []string{"a.go", "b.go", "c.tmp", "a.gif", "m.swp", "n.swf"}
	for _, d := range dirs {
		if err := os.Mkdir(dir+"/"+d, 0777); err != nil {
			t.Error(err)
			return
		} else {
			for _, f := range files {
				if _, err := os.Create(dir + "/" + d + "/" + f); err != nil {
					t.Error(err)
					return
				}
			}
		}
	}

	if gofiles, err := SearchFile(dir, ".go$"); err != nil {
		t.Error(err)
		return
	} else {
		if len(gofiles) != 2*len(dirs) {
			t.Errorf("found %d .go files, expect: %d", len(gofiles), 2*len(dirs))
			return
		} else {
			t.Logf("found %d .go files", len(gofiles))
		}
	}

	if swffiles, err := SearchFile(dir, ".swf$"); err != nil {
		t.Error(err)
		return
	} else {
		if len(swffiles) != 1*len(dirs) {
			t.Errorf("found %d .swf files, expect: %d", len(swffiles), 2*len(dirs))
			return
		} else {
			t.Logf("found %d .swf files", len(swffiles))
		}
	}
}
