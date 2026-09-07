package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
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
	Name        string
	Description string
	Text        []byte
	Properties  map[Property]string
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
		NewLicense("MIT", "", licenseMIT, []Property{PropertyYear, PropertyAuthor}),
		NewLicense("MPL 2.0", "", licenseMPL2, nil),
		NewLicense("GPLv3", "", licenseGPL3, []Property{PropertyYear, PropertyAuthor, PropertyName, PropertyDescription}),
		NewLicense("AGPLv3", "", licenseAGPL3, []Property{PropertyYear, PropertyAuthor, PropertyName, PropertyDescription}),
		NewLicense("LGPLv3", "", licenseLGPL3, nil),
		NewLicense("Apache 2.0", "", licenseApache2, []Property{PropertyYear, PropertyAuthor}),
	}
)

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

func NewLicense(name, description string, text []byte, properties []Property) License {
	propertyMap := make(map[Property]string, len(properties))

	for _, property := range properties {
		propertyMap[property] = ""
	}

	return License{
		Name:        name,
		Description: description,
		Text:        text,
		Properties:  propertyMap,
	}
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
