package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"sync"

	"github.com/database64128/tfo-go"
)

var writeGarbage = flag.Bool("writeGarbage", false, "Write garbage before relaying")

func runServer(ctrlCh chan struct{}) error {
	l, err := tfo.Listen("tcp", ":10240")
	if err != nil {
		return err
	}
	tl := l.(*net.TCPListener)
	defer tl.Close()

	fmt.Println("Started server")
	ctrlCh <- struct{}{}

	tc, err := tl.AcceptTCP()
	if err != nil {
		return err
	}
	defer tc.Close()

	if _, err := io.Copy(io.Discard, tc); err != nil {
		return err
	}

	fmt.Println("Server finished reading")

	if _, err := tc.Write([]byte{'b', 'a', 'r'}); err != nil {
		return err
	}

	fmt.Println("Server wrote bar")

	if err := tc.CloseWrite(); err != nil {
		return err
	}

	return nil
}

func runRelay(ctrlCh chan struct{}) error {
	l, err := tfo.Listen("tcp", ":10241")
	if err != nil {
		return err
	}
	tl := l.(*net.TCPListener)
	defer tl.Close()

	fmt.Println("Started relay")
	ctrlCh <- struct{}{}

	clientConn, err := tl.AcceptTCP()
	if err != nil {
		return err
	}
	defer clientConn.Close()

	c, err := tfo.Dial("tcp", "[::1]:10240")
	if err != nil {
		return err
	}
	remoteConn := c.(*net.TCPConn)
	defer remoteConn.Close()

	fmt.Println("Relay connected")

	if *writeGarbage {
		if _, err := remoteConn.Write([]byte{'b', 'a', 'r'}); err != nil {
			return err
		}
		fmt.Println("Relay wrote garbage")
	}

	var (
		l2rErr  error
		l2rSErr error
		r2lErr  error
		r2lSErr error
	)

	ch := make(chan struct{})

	go func() {
		_, l2rErr = io.Copy(remoteConn, clientConn)
		l2rSErr = remoteConn.CloseWrite()
		ch <- struct{}{}
	}()

	_, r2lErr = io.Copy(clientConn, remoteConn)
	r2lSErr = clientConn.CloseWrite()
	<-ch

	switch {
	case l2rErr != nil:
		return l2rErr
	case l2rSErr != nil:
		return l2rSErr
	case r2lErr != nil:
		return r2lErr
	case r2lSErr != nil:
		return r2lSErr
	default:
		return nil
	}
}

func runClient() error {
	c, err := tfo.Dial("tcp", "[::1]:10241")
	if err != nil {
		return err
	}
	tc := c.(*net.TCPConn)
	defer tc.Close()

	if _, err := tc.Write([]byte{'f', 'o', 'o'}); err != nil {
		return err
	}

	fmt.Println("Client wrote foo")

	if err := tc.CloseWrite(); err != nil {
		return err
	}

	if _, err := io.Copy(io.Discard, tc); err != nil {
		return err
	}

	return nil
}

func main() {
	flag.Parse()

	var wg sync.WaitGroup

	wg.Add(2)

	ctrlCh := make(chan struct{})

	go func() {
		if err := runServer(ctrlCh); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("Server done")
		wg.Done()
	}()

	go func() {
		if err := runRelay(ctrlCh); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("Relay done")
		wg.Done()
	}()

	<-ctrlCh
	<-ctrlCh

	if err := runClient(); err != nil {
		fmt.Println(err)
	}

	fmt.Println("Client done")

	wg.Wait()
}
