package tools

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

const SCAN_INTERVAL = time.Hour * 24

// StartScanner starts a scanner which scans a directory and deletes
// files that are older than maxAge
func StartScanner(directory, maxAge string) error {
	for {
		err := scanDirectory(directory, maxAge)
		if err != nil {
			return err
		}

		time.Sleep(SCAN_INTERVAL)
	}
}

func scanDirectory(directory, maxAge string) error {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return err
	}

	maxDur, err := time.ParseDuration(maxAge)
	if err != nil {
		return err
	}

	for _, file := range files {
		path := filepath.Join(directory, file.Name())
		fi, err := os.Stat(path)
		if err != nil {
			return err
		}

		if time.Since(fi.ModTime()) >= maxDur {
			err = os.Remove(path)
			if err != nil {
				return err
			}
			fmt.Printf("deleted file %v!\n", path)
		}
	}

	return nil
}
