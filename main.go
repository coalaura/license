package main

import (
	"os"

	"github.com/coalaura/plain"
)

var log = plain.New()

func main() {
	err := run(log)
	if err == nil {
		return
	}

	log.Errorln(err)

	os.Exit(1)
}
