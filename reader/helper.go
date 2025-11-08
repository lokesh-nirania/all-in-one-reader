package reader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GetUniqueFilePath(originalPath string) (string, error) {
	dir := filepath.Dir(originalPath)
	base := filepath.Base(originalPath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	path := originalPath
	count := 1

	for {
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			return path, nil
		}
		if err != nil {
			return "", err
		}

		newName := fmt.Sprintf("%s_%d%s", name, count, ext)
		path = filepath.Join(dir, newName)
		count++
	}
}

func HumanizeReadableSize(bytes int64) (float64, string) {
	const (
		_          = iota
		KB float64 = 1 << (10 * iota)
		MB
		GB
		TB
	)
	val := float64(bytes)
	switch {
	case val >= TB:
		return val / TB, "TB"
	case val >= GB:
		return val / GB, "GB"
	case val >= MB:
		return val / MB, "MB"
	case val >= KB:
		return val / KB, "KB"
	default:
		return val, "bytes"
	}
}

func NotifyProgress(n int64, totalSize int64) {
	dlVal, dlUnit := HumanizeReadableSize(n)
	if totalSize > 0 {
		totalVal, totalUnit := HumanizeReadableSize(totalSize)
		percent := float64(n) / float64(totalSize) * 100
		fmt.Printf("\rDownloaded %.2f %s / %.2f %s (%.2f%%)...", dlVal, dlUnit, totalVal, totalUnit, percent)
	} else {
		if dlUnit == "bytes" {
			fmt.Printf("\rDownloaded %.0f bytes...", dlVal)
		} else {
			fmt.Printf("\rDownloaded %.2f %s...", dlVal, dlUnit)
		}
	}
}
