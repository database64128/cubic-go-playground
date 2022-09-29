package main

import (
	"io"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/database64128/tfo-go/v2"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	for {
		stressTFO()
		time.Sleep(time.Second)
	}
}

func stressTFO() {
	netConn, err := tfo.Dial("tcp", "[::1]:1080", []byte{5, 2, 1, 2})
	if err != nil {
		log.Fatal(err)
	}
	conn := netConn.(*net.TCPConn)
	defer conn.Close()

	err = conn.CloseWrite()
	if err != nil {
		log.Println(err)
	}

	b, err := io.ReadAll(conn)
	if err != nil {
		log.Println(err)
	}
	log.Println(b)
}
