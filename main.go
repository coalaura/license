package main

import (
	"os"

	"github.com/coalaura/plain"
)

var Version = "dev"

var log = plain.New()

func main() {
	err := run(log, os.Args[1:])
	if err == nil {
		return
	}

	log.Errorln(err)

	os.Exit(1)
}
