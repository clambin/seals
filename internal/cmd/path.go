package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func escapes(path string) bool {
	return strings.HasPrefix(path, ".."+string(os.PathSeparator))
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
