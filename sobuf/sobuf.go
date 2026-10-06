package main

import (
	"flag"
	"os"
)

var (
	size  int
	force bool
)

func init() {
	flag.IntVar(&size, "s", 0, "Set socket buffer sizes")
	flag.BoolVar(&force, "f", false, "Force setting socket buffer sizes (SO_{SND,RCV}BUFFORCE) on Linux")
}

func main() {
	flag.Parse()

	if flag.NArg() > 0 {
		flag.Usage()
		os.Exit(2)
	}

	os.Exit(run())
}
