package main

import (
	"log"
	"strings"
)

func fatalIfError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Pid's can contain a ":" character which we want to replace
func normalizeId(id string) string {
	return strings.ReplaceAll(id, ":", "-")
}

//
// end of file
//
