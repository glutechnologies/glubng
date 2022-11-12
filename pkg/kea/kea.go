package kea

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

type KeaSocket struct {
	filename string
	listener net.Listener
	_stop    chan bool
	_wg      sync.WaitGroup
}

func handleConection(k *KeaSocket, conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 2048)

ReadLoop:
	for {
		select {
		case <-k._stop:
			return
		default:
			conn.SetDeadline(time.Now().Add(200 * time.Millisecond))
			n, err := conn.Read(buf)
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					continue ReadLoop
				} else if err != io.EOF {
					log.Println("read error", err)
					return
				}
			}
			if n == 0 {
				return
			}
			// Process data from Kea
			log.Printf("received from %v: %s", conn.RemoteAddr(), string(buf[:n]))
		}
	}
}

func runUnixSocketServer(k *KeaSocket) {
	// https://eli.thegreenplace.net/2020/graceful-shutdown-of-a-tcp-server-in-go/
	defer k._wg.Done()

	for {
		// Accept new connections, dispatching them to echoServer
		// in a goroutine.
		conn, err := k.listener.Accept()
		if err != nil {
			select {
			case <-k._stop:
				fmt.Println("Closing socket...")
				k.listener.Close()
				return
			default:
			}
			log.Fatal("accept error:", err)
		} else {
			k._wg.Add(1)
			go func() {
				handleConection(k, conn)
				k._wg.Done()
			}()
		}
	}
}

func (k *KeaSocket) Init(filename string) {
	k.filename = filename
	k._stop = make(chan bool)

	if err := os.RemoveAll(filename); err != nil {
		log.Fatal(err)
	}

	var err error
	k.listener, err = net.Listen("unix", filename)

	if err != nil {
		log.Fatal("listen error:", err)
	}

	// Add one level to WaitGroup
	k._wg.Add(1)
	go runUnixSocketServer(k)
}

func (k *KeaSocket) Close() {
	close(k._stop)
	k.listener.Close()
	k._wg.Wait()
}
