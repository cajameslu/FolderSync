package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"
)

func fileInfo(fileName string) (bool, *time.Time) {
	file, err := os.Stat(fileName)

	if err == nil {
		modTime := file.ModTime()
		return true, &modTime
	}

	return false, nil
}

func isDir(path string) bool {
	file, err := os.Stat(path)
	isFolder := false

	if err == nil {
		isFolder = file.IsDir()
	}

	return isFolder
}

func fileCopy(srcFile string, destFile string) (bool, string) {

	from, err := os.Open(srcFile)
	if err != nil {
		return false, fmt.Sprintf("Source file open error: %s", err)
	}

	defer from.Close()

	to, err := os.OpenFile(destFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return false, fmt.Sprintf("Detination file creating error: %s", err)
	}

	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		return false, fmt.Sprintf("Copy file error: %s", err)
	}

	return true, ""
}

func listFiles(folderName string) (map[string]*time.Time, string) {
	files := map[string]*time.Time{}

	fs, err := ioutil.ReadDir(folderName)

	if err != nil {
		return files, fmt.Sprintf("%s", err)
	}

	for _, f := range fs {
		if !f.IsDir() {
			modTime := f.ModTime()
			files[f.Name()] = &modTime
		}
	}

	return files, ""
}
