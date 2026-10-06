package main

import (
	"flag"
	"os"
)

var size int

func init() {
	flag.IntVar(&size, "size", 0, "Set socket buffer sizes")
}

func main() {
	flag.Parse()

	if flag.NArg() > 0 {
		flag.Usage()
		os.Exit(2)
	}

	os.Exit(run())
}
