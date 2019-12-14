package util

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"io/ioutil"
	"os"
	"regexp"
	"sort"
)

// MatchPath finds the first filename matching a regexp in dir
func MatchPath(dir string, re *regexp.Regexp) (string, error) {
	matches, err := MatchPaths(dir, re)
	if err != nil {
		return "", err
	}
	if len(matches) < 1 {
		return "", os.ErrNotExist
	}
	sort.Strings(matches)
	return matches[0], nil
}

// MatchPaths finds all filenames matching a regexp in dir (non-recursive)
func MatchPaths(dir string, re *regexp.Regexp) ([]string, error) {

	dirEntries, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var matches []string
	for _, entry := range dirEntries {
		if re.MatchString(entry.Name()) {
			matches = append(matches, entry.Name())
		}
	}

	return matches, nil
}

// PathExists is a simple existence check function
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
