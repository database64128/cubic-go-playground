package main

import (
	"crypto/rand"
	"flag"
	"log"

	"go.uber.org/zap"
)

type myStringer struct {
	id     string
	buf    [16]byte
	logger *zap.Logger
}

func (ms *myStringer) String() string {
	_, err := rand.Read(ms.buf[:])
	if err != nil {
		panic(err)
	}

	ms.logger.Info("MyStringer.String() called", zap.String("id", ms.id), zap.Binary("buf", ms.buf[:]))
	return ms.id
}

var verbose = flag.Bool("verbose", false, "Enable verbose logging")

func main() {
	flag.Parse()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	var (
		msser = myStringer{id: "zap.Stringer", logger: logger}
		mss   = myStringer{id: "zap.String", logger: logger}
	)

	if *verbose {
		logger.Debug("Using zap.Stringer", zap.Stringer("msser", &msser))
		logger.Debug("Using zap.String", zap.String("mss", mss.String()))
	}
}
