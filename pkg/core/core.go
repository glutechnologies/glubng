package core

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/glutechnologies/glubng/pkg/kea"
	"github.com/glutechnologies/glubng/pkg/vpp"
)

type Config struct {
	SrcKeaSocket string
	SrcVppSocket string
}

type Core struct {
	control  chan os.Signal
	config   Config
	sessions Sessions
	vpp      vpp.Client
	kea      kea.KeaSocket
	wg       sync.WaitGroup
}

func (c *Core) LoadConfig() {
	// Load configuration
	configSrc := flag.String("config", "/etc/glubng.toml", "Config source path")
	flag.Parse()

	body, err := os.ReadFile(*configSrc)

	if err != nil {
		log.Fatalf("Error loading configuration file")
	}

	_, err = toml.Decode(string(body), &c.config)

	if err != nil {
		log.Fatalf("Error decoding configuration file")
	}
}

func (c *Core) Init() {
	// Load initial configuration
	c.LoadConfig()

	// Init kea listener
	c.kea.Init(c.config.SrcKeaSocket)

	// Init VPP
	c.vpp.Init(c.config.SrcVppSocket, true)

	// Init Sessions
	c.sessions.Init(&c.vpp)

	// Create a channel to process signals
	c.control = make(chan os.Signal, 1)
	signal.Notify(c.control, syscall.SIGINT, syscall.SIGTERM)

	// Process messages received from Kea DHCP Server
	c.wg.Add(1)
	go c.ProcessKeaMessages()

	fmt.Println("Running GluBNGd...")

	// Add 1 to wg counter
	c.wg.Add(1)
	go func() {
		<-c.control
		c.kea.Close()
		c.wg.Done()
	}()

	c.wg.Wait()
	fmt.Println("Exiting GluBNGd...")
}

func (c *Core) ProcessKeaMessages() {
	for {
		select {
		case <-c.control:
			// Receive CONTROL-C exit goroutine
			goto endLoop
		case msg := <-c.kea.Message:
			switch msg.Callout {
			case kea.CALLOUT_LEASE4_SELECT:
				// New Lease selected
				iface, err := strconv.ParseInt(msg.Query.Option82CID, 16, 0)
				if err != nil {
					// Error parsing circuit-id
					log.Printf("Error in ProcessKeaMessages, %s", err.Error())
					break
				}
				ses := &Session{Iface: int(iface), IPv4: msg.Lease.Address}
				c.sessions.AddSession(ses)
			}
		}
	}
endLoop:
	c.wg.Done()
}
