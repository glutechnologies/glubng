package vpp

import (
	"log"
	"net"

	"go.fd.io/govpp"
	"go.fd.io/govpp/binapi/fib_types"
	"go.fd.io/govpp/binapi/ip"
	"go.fd.io/govpp/binapi/ip_types"
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

func (c *Client) AddSession(ipv4 string, iface uint32) {
	log.Println("Add session to VPP")

	goip := net.ParseIP(ipv4)

	if goip == nil {
		log.Println("Error adding session in parse ip")
		return
	}

	vppip := ip_types.Address{
		Af: ip_types.ADDRESS_IP4,
		Un: ip_types.AddressUnionIP4(ip_types.IP4Address{
			goip[12], goip[13], goip[14], goip[15],
		}),
	}

	c.addRouteToVPP(&vppip, iface)
}

func (c *Client) RemoveSession() {
	log.Println("Remove session from VPP")
}

func (c *Client) addRouteToVPP(ipv4 *ip_types.Address, iface uint32) error {
	ch, err := c.conn.NewAPIChannel()
	if err != nil {
		log.Println("ERROR: creating channel failed:", err)
		return err
	}
	defer ch.Close()

	path := fib_types.FibPath{SwIfIndex: iface}

	req := &ip.IPRouteAddDel{IsAdd: true,
		Route: ip.IPRoute{TableID: 0,
			Prefix: ip_types.Prefix{Address: *ipv4, Len: 32},
			Paths:  []fib_types.FibPath{path}}}

	reply := &ip.IPRouteAddDelReply{}

	if err = ch.SendRequest(req).ReceiveReply(reply); err != nil {
		log.Println("Error adding route", err)
		return err
	}

	return nil
}
