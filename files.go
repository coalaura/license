package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func FindLicense() (string, bool, error) {
	dir, err := os.Open(".")
	if err != nil {
		return "", false, err
	}

	defer dir.Close()

	for {
		files, err := dir.ReadDir(128)

		for _, file := range files {
			name := file.Name()

			if !isLicenseFile(name) {
				continue
			}

			if file.Type().IsRegular() {
				return name, true, nil
			}
		}

		if errors.Is(err, io.EOF) {
			return "", false, nil
		}

		if err != nil {
			return "", false, err
		}
	}
}

func isLicenseFile(name string) bool {
	if strings.EqualFold(name, "license") {
		return true
	}

	ext := filepath.Ext(name)
	if !strings.EqualFold(ext, ".md") && !strings.EqualFold(ext, ".txt") {
		return false
	}

	stem := name[:len(name)-len(ext)]

	base, language, localized := strings.Cut(stem, ".")
	if !strings.EqualFold(base, "license") {
		return false
	}

	return !localized || isLanguageTag(language)
}

func isLanguageTag(tag string) bool {
	first := true

	for subtag := range strings.SplitSeq(tag, "-") {
		if len(subtag) == 0 || len(subtag) > 8 {
			return false
		}

		for i := range len(subtag) {
			char := subtag[i]

			if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' {
				continue
			}

			if first || char < '0' || char > '9' {
				return false
			}
		}

		first = false
	}

	return !first
}
