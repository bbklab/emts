package inc

import (
	"fmt"
	"os"
	"regexp"
)

// SearchFile() search direcotry for specified extension files
func SearchFile(path string, ext string) (files []string, err error) {

	dirinfo, err := os.Stat(path)
	if err != nil {
		return []string{}, err
	}

	if !dirinfo.IsDir() {
		return []string{}, fmt.Errorf("%s isn't directory!", path)
	}

	dir, err := os.Open(path)
	if err != nil {
		return []string{}, fmt.Errorf("%s open failed: ", err.Error())
	}
	defer dir.Close()

	items, err := dir.Readdir(-1)
	if err != nil {
		return []string{}, fmt.Errorf("%s readdir failed: ", err.Error())
	}
	if path[len(path)-1] != '/' {
		path += "/"
	}
	for _, item := range items {
		if item.IsDir() {
			newpath := path + item.Name()
			if temp, err := SearchFile(newpath, ext); err == nil {
				for _, v := range temp {
					files = append(files, v)
				}
			}
		}
		if matched, err := regexp.MatchString(ext, item.Name()); matched && err == nil {
			files = append(files, path+item.Name())
		}
	}
	return files, nil
}
