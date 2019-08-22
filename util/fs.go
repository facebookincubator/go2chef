package util

import (
	"io/ioutil"
	"os"
	"regexp"
	"sort"
)

func MatchFile(dir string, re *regexp.Regexp) (string, error) {
	dirEntries, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", err
	}

	matches := make([]string, 0)
	for _, entry := range dirEntries {
		if re.MatchString(entry.Name()) {
			matches = append(matches, entry.Name())
		}
	}

	if len(matches) < 1 {
		return "", os.ErrNotExist
	}

	sort.Strings(matches)

	return matches[0], nil
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
