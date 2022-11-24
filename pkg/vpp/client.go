package vpp

import (
	"log"

	"go.fd.io/govpp"
	"go.fd.io/govpp/core"
)

// VPP related configuration
type VPPConfig struct {
	SrcVPPSocket        string
	UplinkInterfaceName string
	UplinkInterfaceIPv4 string
}

type Client struct {
	config VPPConfig
	conn   *core.Connection
}

func (c *Client) Init(config *VPPConfig) {
	// Initialize all struct members
	c.config = *config

	conn, connEv, err := govpp.AsyncConnect(c.config.SrcVPPSocket, core.DefaultMaxReconnectAttempts, core.DefaultReconnectInterval)
	if err != nil {
		log.Fatalln("ERROR:", err)
	}

	c.conn = conn
	// wait for Connected event
	e := <-connEv
	if e.State != core.Connected {
		log.Fatalln("ERROR: connecting to VPP failed:", e.Error)
	}

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
