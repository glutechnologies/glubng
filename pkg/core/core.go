package core

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/glutechnologies/glubng/pkg/kea"
	"github.com/glutechnologies/glubng/pkg/vpp"
)

// Configuration aggregation
type CoreConfig struct {
	Misc MiscConfig    `toml:"misc"`
	Vpp  vpp.VPPConfig `toml:"vpp"`
}

type MiscConfig struct {
	SrcKeaSocket string
}

type Core struct {
	ifacesFile string
	configFile string
	control    chan os.Signal
	config     CoreConfig
	sessions   Sessions
	vpp        vpp.Client
	kea        kea.KeaSocket
	wg         sync.WaitGroup
}

func (c *Core) LoadConfig() {
	body, err := os.ReadFile(c.configFile)

	if err != nil {
		log.Fatalf("Error loading configuration file")
	}

	_, err = toml.Decode(string(body), &c.config)

	if err != nil {
		log.Fatalf("Error decoding configuration file")
	}
}

func (c *Core) WriteConfig() {
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(c.config)

	if err != nil {
		log.Fatalf("Error encoding configuration in TOML")
	}

	err = os.WriteFile(c.configFile, buf.Bytes(), 0666)

	if err != nil {
		log.Fatalf("Error writing configuration file")
	}
}

func (c *Core) Init() {
	// Define flags
	configFile := flag.String("config", "/etc/glubng.toml", "Config source path")
	ifacesFile := flag.String("interfaces", "/etc/interfaces.toml", "Config interfaces source path")
	flag.Parse()

	// Store files
	c.configFile = *configFile
	c.ifacesFile = *ifacesFile

	// Load initial configuration
	c.LoadConfig()

	// Init kea listener
	c.kea.Init(c.config.Misc.SrcKeaSocket)

	// Init VPP
	c.vpp.Init(&c.config.Vpp, c.ifacesFile)

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
		c.vpp.Close()
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
			case kea.CALLOUT_LEASE4_SELECT, kea.CALLOUT_LEASE4_RENEW:
				// New Lease selected
				iface, err := ConvertCIDToInt(msg.Query.Option82CID)
				if err != nil {
					// Error parsing circuit-id
					log.Printf("Error in ProcessKeaMessages, %s", err.Error())
					break
				}
				// ParseIP is an slice[16], positions 12,13,14,15 are used for IPv4
				goip := net.ParseIP(msg.Lease.Address)

				if goip == nil {
					log.Println("Error adding session in parse ip")
					break
				}
				ses := &Session{Iface: int(iface), IPv4: goip}
				c.sessions.AddSession(ses)
			case kea.CALLOUT_LEASE4_RELEASE, kea.CALLOUT_LEASE4_EXPIRE:
				// Remove Session when a lease expires
				c.sessions.RemoveSession(msg.Lease.Address)
			}
		}
	}
endLoop:
	c.wg.Done()
}
