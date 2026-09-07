package main

import (
	"os"

	"github.com/coalaura/logger/plain"
)

var log = plain.New()

func main() {
	existing, found, err := FindLicense()
	log.MustFail(err)

	if found {
		log.Errorf("license file %q already exists\n", existing)

		os.Exit(2)
	}

	test, err := os.OpenFile("LICENSE", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	log.MustFail(err)

	defer test.Close()

	license := licenses[0]

	license.Properties[PropertyYear] = "2026"
	license.Properties[PropertyAuthor] = "coalaura"

	err = license.Write(test)
	log.MustFail(err)
}
