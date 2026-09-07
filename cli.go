package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/coalaura/plain"
)

type Options struct {
	License     string
	Author      string
	Year        string
	Name        string
	Description string
}

func (options Options) property(property Property) string {
	switch property {
	case PropertyYear:
		return options.Year
	case PropertyAuthor:
		return options.Author
	case PropertyName:
		return options.Name
	case PropertyDescription:
		return options.Description
	default:
		return ""
	}
}

func run(options Options) error {
	directory, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	return generateLicense(options, directory, time.Now())
}

func generateLicense(options Options, directory string, now time.Time) error {
	existing, found, err := findExistingLicense(directory)
	if err != nil {
		return fmt.Errorf("look for an existing license: %w", err)
	}

	if found {
		prompt := fmt.Sprintf("Replace existing %q?", existing)

		confirmed, err := log.Confirm(prompt, false)
		if err != nil {
			return fmt.Errorf("confirm license replacement: %w", err)
		}

		if !confirmed {
			return nil
		}
	}

	license, err := promptForLicense(options, directory, now)
	if err != nil {
		return err
	}

	var contents bytes.Buffer

	err = license.Write(&contents)
	if err != nil {
		return fmt.Errorf("render license: %w", err)
	}

	path := filepath.Join(directory, "LICENSE")

	var sameFile bool

	if found {
		existingPath := filepath.Join(directory, existing)

		existingInfo, err := os.Stat(existingPath)
		if err != nil {
			return fmt.Errorf("inspect existing license: %w", err)
		}

		targetInfo, statErr := os.Stat(path)
		if statErr == nil {
			sameFile = os.SameFile(existingInfo, targetInfo)
		} else if !os.IsNotExist(statErr) {
			return fmt.Errorf("inspect LICENSE target: %w", statErr)
		}
	}

	flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL

	if sameFile {
		flags = os.O_WRONLY | os.O_TRUNC
	}

	file, err := os.OpenFile(path, flags, 0644)
	if err != nil {
		return fmt.Errorf("create LICENSE: %w", err)
	}

	_, writeErr := file.Write(contents.Bytes())
	closeErr := file.Close()

	if writeErr != nil {
		_ = os.Remove(path)

		return fmt.Errorf("write LICENSE: %w", writeErr)
	}

	if closeErr != nil {
		_ = os.Remove(path)

		return fmt.Errorf("close LICENSE: %w", closeErr)
	}

	if found && !sameFile {
		existingPath := filepath.Join(directory, existing)

		err = os.Remove(existingPath)
		if err != nil {
			_ = os.Remove(path)

			return fmt.Errorf("remove old license: %w", err)
		}
	}

	log.Println("Wrote LICENSE.")

	return nil
}

func promptForLicense(settings Options, directory string, now time.Time) (License, error) {
	var license License

	if settings.License != "" {
		var found bool

		license, found = findLicense(settings.License)
		if !found {
			return License{}, fmt.Errorf("unknown license %q", settings.License)
		}
	} else {
		options := make([]plain.SelectOption, len(licenses))

		for index, available := range licenses {
			options[index] = available
		}

		index, err := log.SelectWithDescription("Choose a license: ", options)
		if err != nil {
			return License{}, fmt.Errorf("select license: %w", err)
		}

		if index < 0 || index >= len(licenses) {
			return License{}, fmt.Errorf("select license: invalid selection %d", index)
		}

		license = licenses[index]
	}

	license.Properties = make(map[Property]string, len(license.RequiredProperties))

	for _, property := range license.RequiredProperties {
		value := settings.property(property)
		if value != "" {
			license.Properties[property] = value

			continue
		}

		value, err := promptForProperty(property, directory, now)
		if err != nil {
			return License{}, err
		}

		license.Properties[property] = value
	}

	return license, nil
}

func promptForProperty(property Property, directory string, now time.Time) (string, error) {
	name := propertyName(property)
	defaultValue := defaultProperty(property, directory, now)

	prompt := fmt.Sprintf("%s: ", title(name))

	if defaultValue != "" {
		prompt = fmt.Sprintf("%s [%s]: ", title(name), defaultValue)
	}

	for {
		value, err := log.Read(prompt, 512)
		if err != nil {
			return "", fmt.Errorf("read %s: %w", name, err)
		}

		value = strings.TrimSpace(value)
		if value != "" {
			return value, nil
		}

		if defaultValue != "" {
			return defaultValue, nil
		}
	}
}

func defaultProperty(property Property, directory string, now time.Time) string {
	switch property {
	case PropertyYear:
		return strconv.Itoa(now.Year())
	case PropertyAuthor:
		return repositoryOwner(directory)
	case PropertyName:
		return filepath.Base(directory)
	default:
		return ""
	}
}

func repositoryOwner(directory string) string {
	command := exec.Command("git", "config", "--get", "remote.origin.url")
	command.Dir = directory

	output, err := command.Output()
	if err != nil {
		return ""
	}

	return ownerFromRemote(string(output))
}

func ownerFromRemote(remote string) string {
	remote = strings.TrimSpace(remote)
	remote = strings.ReplaceAll(remote, "\\", "/")
	remote = strings.TrimSuffix(remote, "/")

	colon := strings.LastIndexByte(remote, ':')
	slash := strings.IndexByte(remote, '/')

	scpStyle := colon >= 0 && (slash == -1 || colon < slash)
	if scpStyle {
		remote = remote[colon+1:]
	}

	parts := strings.Split(remote, "/")
	if len(parts) < 2 {
		return ""
	}

	return parts[len(parts)-2]
}

func propertyName(property Property) string {
	switch property {
	case PropertyYear:
		return "year"
	case PropertyAuthor:
		return "author"
	case PropertyName:
		return "project"
	case PropertyDescription:
		return "description"
	default:
		return "property"
	}
}

func title(value string) string {
	if value == "" {
		return ""
	}

	return strings.ToUpper(value[:1]) + value[1:]
}
