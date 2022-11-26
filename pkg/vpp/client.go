package vpp

import (
	"log"

	"go.fd.io/govpp"
	"go.fd.io/govpp/core"
)

type Client struct {
	config     VPPConfig
	ifaces     map[string]Iface
	ifacesFile string
	conn       *core.Connection
}

func (c *Client) Init(config *VPPConfig, ifacesFile string) {
	// Initialize all struct members
	c.config = *config
	c.ifacesFile = ifacesFile

	conn, connEv, err := govpp.AsyncConnect(c.config.SrcVPPSocket, core.DefaultMaxReconnectAttempts, core.DefaultReconnectInterval)
	if err != nil {
		log.Fatalln("Async connect to VPP", err)
	}

	c.conn = conn
	// wait for Connected event
	e := <-connEv
	if e.State != core.Connected {
		log.Fatalln("Connecting to VPP failed:", e.Error)
	}

	// Load CPE Interface configurations
	c.LoadIfacesConfig()
}

func (c *Client) Close() {
	c.conn.Disconnect()
}

func (c *Client) AddSession() {
	log.Println("Add session to VPP")
}

func (c *Client) RemoveSession() {
	log.Println("Remove session from VPP")
}
