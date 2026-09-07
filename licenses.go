package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"strings"
)

type Property int

const (
	PropertyInvalid Property = iota
	PropertyYear
	PropertyAuthor
	PropertyName
	PropertyDescription
)

type License struct {
	Name               string
	Summary            string
	Text               []byte
	RequiredProperties []Property
	Properties         map[Property]string
}

var (
	//go:embed licenses/agpl_3.txt
	licenseAGPL3 []byte

	//go:embed licenses/apache_2.txt
	licenseApache2 []byte

	//go:embed licenses/gpl_3.txt
	licenseGPL3 []byte

	//go:embed licenses/lgpl_3.txt
	licenseLGPL3 []byte

	//go:embed licenses/mit.txt
	licenseMIT []byte

	//go:embed licenses/mpl_2.txt
	licenseMPL2 []byte

	licenses = []License{
		NewLicense("MIT", "Simple and permissive, with minimal restrictions on reuse.", licenseMIT, []Property{PropertyYear, PropertyAuthor}),
		NewLicense("Apache 2.0", "Permissive, with an explicit patent grant from contributors.", licenseApache2, []Property{PropertyYear, PropertyAuthor}),
		NewLicense("GPLv3", "Strong copyleft requiring derivatives to remain open source.", licenseGPL3, []Property{PropertyYear, PropertyAuthor, PropertyName, PropertyDescription}),
		NewLicense("LGPLv3", "Library-focused copyleft that permits proprietary linking.", licenseLGPL3, nil),
		NewLicense("AGPLv3", "Strong copyleft that also covers software used over a network.", licenseAGPL3, []Property{PropertyYear, PropertyAuthor, PropertyName, PropertyDescription}),
		NewLicense("MPL 2.0", "File-level copyleft that can be combined with proprietary code.", licenseMPL2, nil),
	}

	licenseAliases = map[string]string{
		"mit": "MIT",

		"apache":    "Apache 2.0",
		"apache2":   "Apache 2.0",
		"apache20":  "Apache 2.0",
		"apachev2":  "Apache 2.0",
		"apachev20": "Apache 2.0",

		"gpl":   "GPLv3",
		"gpl3":  "GPLv3",
		"gplv3": "GPLv3",

		"lgpl":   "LGPLv3",
		"lgpl3":  "LGPLv3",
		"lgplv3": "LGPLv3",

		"agpl":   "AGPLv3",
		"agpl3":  "AGPLv3",
		"agplv3": "AGPLv3",

		"mpl":    "MPL 2.0",
		"mpl2":   "MPL 2.0",
		"mpl20":  "MPL 2.0",
		"mplv2":  "MPL 2.0",
		"mplv20": "MPL 2.0",
	}
)

func (l License) Label() string {
	return l.Name
}

func (l License) Description() string {
	return l.Summary
}

func (l License) Write(wr io.Writer) error {
	var index int

	for {
		open := bytes.IndexByte(l.Text[index:], '[')
		if open == -1 {
			break
		}

		open += index

		_, err := wr.Write(l.Text[index:open])
		if err != nil {
			return err
		}

		index = open + 1

		closed := bytes.IndexByte(l.Text[index:], ']')
		if closed == -1 {
			continue
		}

		closed += index

		name := l.Text[index:closed]

		index = closed + 1

		property := getPropertyFromName(name)
		if property == PropertyInvalid {
			_, err := wr.Write(l.Text[open:index])
			if err != nil {
				return err
			}

			continue
		}

		value, ok := l.Properties[property]
		if !ok {
			return fmt.Errorf("missing property %q", string(name))
		}

		_, err = wr.Write([]byte(value))
		if err != nil {
			return err
		}
	}

	if index < len(l.Text) {
		_, err := wr.Write(l.Text[index:])
		return err
	}

	return nil
}

func NewLicense(name, summary string, text []byte, properties []Property) License {
	propertyMap := make(map[Property]string, len(properties))

	for _, property := range properties {
		propertyMap[property] = ""
	}

	return License{
		Name:               name,
		Summary:            summary,
		Text:               text,
		RequiredProperties: append([]Property(nil), properties...),
		Properties:         propertyMap,
	}
}

func findLicense(value string) (License, bool) {
	alias := normalizeLicenseName(value)

	name, found := licenseAliases[alias]
	if !found {
		return License{}, false
	}

	for _, license := range licenses {
		if license.Name == name {
			return license, true
		}
	}

	return License{}, false
}

func normalizeLicenseName(value string) string {
	value = strings.ToLower(value)

	var normalized strings.Builder

	normalized.Grow(len(value))

	for _, character := range value {
		letter := character >= 'a' && character <= 'z'
		digit := character >= '0' && character <= '9'

		if letter || digit {
			normalized.WriteRune(character)
		}
	}

	return normalized.String()
}

func getPropertyFromName(name []byte) Property {
	switch len(name) {
	case 4:
		if bytes.Equal(name, []byte("year")) {
			return PropertyYear
		}

		if bytes.Equal(name, []byte("name")) {
			return PropertyName
		}
	case 6:
		if bytes.Equal(name, []byte("author")) {
			return PropertyAuthor
		}
	case 11:
		if bytes.Equal(name, []byte("description")) {
			return PropertyDescription
		}
	}

	return PropertyInvalid
}
