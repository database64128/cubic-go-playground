package main

import (
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/database64128/tfo-go"
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
	netConn, err := tfo.Dial("tcp", "[::1]:1080")
	if err != nil {
		log.Fatal(err)
	}
	conn := netConn.(tfo.Conn)
	defer conn.Close()

	_, err = conn.Write([]byte{5, 2, 1, 2})
	if err != nil {
		log.Println(err)
	}

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
