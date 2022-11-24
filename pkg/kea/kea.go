package kea

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

type KeaSocket struct {
	Filename string
	Listener net.Listener
	Message  chan KeaResult
	stop     chan bool
	wg       sync.WaitGroup
}

func handleConection(k *KeaSocket, conn net.Conn) {
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(200 * time.Millisecond))
	d := json.NewDecoder(conn)
	var env Envelope
	err := d.Decode(&env)
	if err != nil {
		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			return
		} else if err != io.EOF {
			log.Println("read error", err)
			return
		}
	}
	// Process data from Kea
	processDataFromConnection(k, &env)
}

func runUnixSocketServer(k *KeaSocket) {
	// https://eli.thegreenplace.net/2020/graceful-shutdown-of-a-tcp-server-in-go/
	defer k.wg.Done()

	for {
		// Accept new connections, dispatching them to echoServer
		// in a goroutine.
		conn, err := k.Listener.Accept()
		if err != nil {
			select {
			case <-k.stop:
				fmt.Println("Closing socket...")
				k.Listener.Close()
				return
			default:
			}
			log.Fatal("accept error:", err)
		} else {
			k.wg.Add(1)
			go func() {
				handleConection(k, conn)
				k.wg.Done()
			}()
		}
	}
}

func (k *KeaSocket) Init(Filename string) {
	k.Filename = Filename
	k.stop = make(chan bool)
	k.Message = make(chan KeaResult)

	if err := os.RemoveAll(Filename); err != nil {
		log.Fatal(err)
	}

	var err error
	k.Listener, err = net.Listen("unix", Filename)

	if err != nil {
		log.Fatal("listen error:", err)
	}

	// Add one level to WaitGroup
	k.wg.Add(1)
	go runUnixSocketServer(k)
}

func (k *KeaSocket) Close() {
	close(k.stop)
	k.Listener.Close()
	k.wg.Wait()
}
