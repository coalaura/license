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
	Interactive bool
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

		confirmed, err := log.ConfirmWithEcho(prompt, false, " ")
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

	if settings.Interactive {
		name, err := recommendLicense()
		if err != nil {
			return License{}, err
		}

		var found bool

		license, found = findLicense(name)
		if !found {
			return License{}, fmt.Errorf("recommended unknown license %q", name)
		}

		log.Printf("Recommended license: %s - %s\n", license.Name, license.Summary)
	} else if settings.License != "" {
		var found bool

		license, found = findLicense(settings.License)
		if !found {
			return License{}, fmt.Errorf("unknown license %q", settings.License)
		}

		log.Printf("Choose a license: %s\n", license.Name)
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
			prompt := propertyPrompt(property, "")

			log.Printf("%s%s\n", prompt, value)

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

func recommendLicense() (string, error) {
	copyleft, err := log.ConfirmWithEcho("Must modified versions remain open source?", false, " ")
	if err != nil {
		return "", fmt.Errorf("ask about copyleft: %w", err)
	}

	if !copyleft {
		patentGrant, err := log.ConfirmWithEcho("Do you want an explicit patent grant from contributors?", false, " ")
		if err != nil {
			return "", fmt.Errorf("ask about patent protection: %w", err)
		}

		if patentGrant {
			return "apache", nil
		}

		return "mit", nil
	}

	networkUse, err := log.ConfirmWithEcho("Will users interact with the software primarily over a network?", false, " ")
	if err != nil {
		return "", fmt.Errorf("ask about network use: %w", err)
	}

	if networkUse {
		return "agpl", nil
	}

	proprietaryLinking, err := log.ConfirmWithEcho("Is this a library that proprietary applications may link to?", false, " ")
	if err != nil {
		return "", fmt.Errorf("ask about proprietary linking: %w", err)
	}

	if proprietaryLinking {
		return "lgpl", nil
	}

	fileLevel, err := log.ConfirmWithEcho("Should open-source requirements apply only to modified files?", false, " ")
	if err != nil {
		return "", fmt.Errorf("ask about file-level copyleft: %w", err)
	}

	if fileLevel {
		return "mpl", nil
	}

	return "gpl", nil
}

func promptForProperty(property Property, directory string, now time.Time) (string, error) {
	name := propertyName(property)
	defaultValue := defaultProperty(property, directory, now)

	prompt := propertyPrompt(property, defaultValue)

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

func propertyPrompt(property Property, defaultValue string) string {
	name := title(propertyName(property))

	if defaultValue != "" {
		return fmt.Sprintf("%s [%s]: ", name, defaultValue)
	}

	return fmt.Sprintf("%s: ", name)
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
	remote := gitConfig(directory, "remote.origin.url")

	owner := ownerFromRemote(remote)
	if owner != "" {
		return owner
	}

	return gitConfig(directory, "user.name")
}

func gitConfig(directory, key string) string {
	command := exec.Command("git", "config", "--get", key)

	command.Dir = directory

	output, err := command.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
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
