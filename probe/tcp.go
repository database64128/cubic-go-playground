package main

import (
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	address = flag.String("address", "", "Target address in host:port form")
	size    = flag.Int("size", 1, "Size of random payload to send")
)

func main() {
	flag.Parse()

	if *address == "" {
		fmt.Println("Missing target address: -address <address>.")
		flag.Usage()
		return
	}

	if *size < 0 {
		fmt.Println("Payload size cannot be less than zero.")
		return
	}

	conn, err := net.Dial("tcp", *address)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	log.Printf("Connected %s -> %s", conn.LocalAddr(), conn.RemoteAddr())

	if *size > 0 {
		buf := make([]byte, *size)
		rand.Read(buf)

		_, err = conn.Write(buf)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Written %d byte(s), the first byte is %v", *size, buf[0])
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Printf("Received %s, stopping...", sig.String())
		conn.SetDeadline(time.Now())
	}()

	buf := make([]byte, 65535)
	n, err := io.ReadFull(conn, buf)
	if err != nil && !errors.Is(err, os.ErrDeadlineExceeded) {
		log.Println(err)
	}
	log.Printf("%d byte(s) read", n)
}
