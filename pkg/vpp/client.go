package vpp

import (
	"log"

	"go.fd.io/govpp"
	"go.fd.io/govpp/core"
)

type Client struct {
	sockAddr string
	conn     *core.Connection
	enabled  bool
}

func (c *Client) Init(sockAddr string, enabled bool) {
	// Initialize all struct members
	c.sockAddr = sockAddr
	c.enabled = enabled

	// If vpp is not enabled return
	if !c.enabled {
		return
	}

	conn, connEv, err := govpp.AsyncConnect(sockAddr, core.DefaultMaxReconnectAttempts, core.DefaultReconnectInterval)
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
	// If vpp is not enabled return
	if !c.enabled {
		return
	}

	c.conn.Disconnect()
}

func (c *Client) AddSession() {
	log.Println("Add session to VPP")
}

func (c *Client) RemoveSession() {
	log.Println("Remove session from VPP")
}
