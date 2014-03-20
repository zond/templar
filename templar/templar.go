package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/zond/templar"
)

const (
	template = "template"
	blob     = "blob"
)

func main() {
	wd, _ := os.Getwd()
	dir := flag.String("dir", wd, "Where to look for template files")
	dst := flag.String("dst", "", "Where to put the source code version of the templates")
	typ := flag.String("type", "", fmt.Sprintf("%#v or %#v", template, blob))

	flag.Parse()

	if *dst == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *typ == template {
		if err := templar.GenerateTemplates(*dir, *dst); err != nil {
			panic(err)
		}
	} else if *typ == blob {
		if err := templar.GenerateBlobs(*dir, *dst); err != nil {
			panic(err)
		}
	} else {
		flag.Usage()
		os.Exit(2)
	}
}
