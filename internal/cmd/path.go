package cmd

import (
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
)

var getWd = os.Getwd

func makeRelativePath(base string, source string) (string, error) {
	if !filepath.IsAbs(source) {
		cwd, err := getWd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
		source = filepath.Join(cwd, source)
	}
	return filepath.Rel(base, source)
}

func shouldUpdate(source, destination string) (bool, error) {
	sourceFInfo, err := os.Stat(source)
	if err != nil {
		return false, fmt.Errorf("%s does not exist", source)
	}

	destinationFInfo, err := os.Stat(destination)
	if err == nil {
		return sourceFInfo.ModTime().After(destinationFInfo.ModTime()), nil
	}
	return true, nil
}

func isWritableDirectory(path string) error {
	fInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat: %w", err)
	}
	if !fInfo.IsDir() {
		return errors.New("not a directory")
	}
	if err = unix.Access(path, unix.W_OK); err != nil {
		return errors.New("not writable")
	}
	return nil
}
