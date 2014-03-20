package main

import (
	"flag"
	"os"

	"github.com/zond/templar"
)

func main() {
	wd, _ := os.Getwd()
	dir := flag.String("dir", wd, "Where to look for template files")
	dst := flag.String("dst", "", "Where to put the source code version of the templates")

	flag.Parse()

	if *dst == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := templar.Generate(*dir, *dst); err != nil {
		panic(err)
	}
}
