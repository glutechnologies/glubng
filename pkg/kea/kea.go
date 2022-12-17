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

	"github.com/glutechnologies/glubng/pkg/vpp"
)

type KeaSocket struct {
	Filename   string
	Listener   net.Listener
	Message    chan KeaResult
	stop       chan bool
	wg         sync.WaitGroup
	ifacesSwIf map[int]vpp.Iface
}

func (k *KeaSocket) handleConection(conn net.Conn) {
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
	res := processDataFromConnection(k, &env)

	// Send response
	sendResponse(k, res, conn)
}

func (k *KeaSocket) runUnixSocketServer() {
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
				k.handleConection(conn)
				k.wg.Done()
			}()
		}
	}
}

func (k *KeaSocket) Init(filename string, ifaces map[int]vpp.Iface) {
	k.Filename = filename
	k.stop = make(chan bool)
	k.Message = make(chan KeaResult)
	k.ifacesSwIf = ifaces

	if err := os.RemoveAll(filename); err != nil {
		log.Fatal(err)
	}

	var err error
	k.Listener, err = net.Listen("unix", filename)

	if err != nil {
		log.Fatal("listen error:", err)
	}

	// Add one level to WaitGroup
	k.wg.Add(1)
	go k.runUnixSocketServer()
}

func (k *KeaSocket) Close() {
	close(k.stop)
	k.Listener.Close()
	k.wg.Wait()
}
