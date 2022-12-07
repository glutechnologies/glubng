package vpp

import (
	"log"
	"net"
	"net/netip"

	"go.fd.io/govpp"
	"go.fd.io/govpp/api"
	"go.fd.io/govpp/binapi/arp"
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
	ch         api.Channel
	gwLoopSwIf int
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

	c.ch, err = c.conn.NewAPIChannel()
	if err != nil {
		log.Fatalf("Error creating channel failed, %s", err.Error())
	}

	// Load CPE Interface configurations
	c.LoadIfacesConfig()

	// Configure VPP
	c.configProxyArp()
	c.configIPv4GwLoopback()
	c.configCPEInterfaces()
	c.configDHCPRelay()
}

func (c *Client) Close() {
	c.ch.Close()
	c.conn.Disconnect()
}

func (c *Client) AddSession(ipv4 net.IP, iface uint32) {
	log.Printf("Add session to VPP, IPv4: %s, SwIf: %d", ipv4.String(), iface)

	vppip := ip_types.Address{
		Af: ip_types.ADDRESS_IP4,
		Un: ip_types.AddressUnionIP4(ip_types.IP4Address{
			ipv4[12], ipv4[13], ipv4[14], ipv4[15],
		}),
	}

	c.addDelRouteToVPP(&vppip, iface, true)
}

func (c *Client) RemoveSession(ipv4 net.IP, iface uint32) {
	log.Printf("Remove session from VPP, IPv4: %s, SwIf: %d", ipv4.String(), iface)

	vppip := ip_types.Address{
		Af: ip_types.ADDRESS_IP4,
		Un: ip_types.AddressUnionIP4(ip_types.IP4Address{
			ipv4[12], ipv4[13], ipv4[14], ipv4[15],
		}),
	}

	c.addDelRouteToVPP(&vppip, iface, false)
}

func (c *Client) addDelRouteToVPP(ipv4 *ip_types.Address, iface uint32, isAdd bool) error {
	path := fib_types.FibPath{SwIfIndex: iface}

	req := &ip.IPRouteAddDel{IsAdd: isAdd,
		Route: ip.IPRoute{TableID: 0,
			Prefix: ip_types.Prefix{Address: *ipv4, Len: 32},
			Paths:  []fib_types.FibPath{path}}}

	reply := &ip.IPRouteAddDelReply{}

	if err := c.ch.SendRequest(req).ReceiveReply(reply); err != nil {
		log.Println("Error adding route", err)
		return err
	}

	return nil
}

func (c *Client) configProxyArp() {
	// Configure ProxyArp
	for _, v := range c.config.IPv4Pool {
		net, err := netip.ParsePrefix(v)

		if err != nil {
			log.Fatalf("Error parsing IPv4Pool, %s", err.Error())
		}

		// Very bad code :(
		var last netip.Addr
		for last = net.Addr(); net.Contains(last); last = last.Next() {
		}
		last = last.Prev()

		first := net.Addr()
		req := &arp.ProxyArpAddDel{IsAdd: true,
			Proxy: arp.ProxyArp{
				TableID: 0,
				Low:     ip_types.IP4Address{first.As4()[0], first.As4()[1], first.As4()[2], first.As4()[3]},
				Hi:      ip_types.IP4Address{last.As4()[0], last.As4()[1], last.As4()[2], last.As4()[3]},
			}}

		reply := &arp.ProxyArpAddDelReply{}

		if err = c.ch.SendRequest(req).ReceiveReply(reply); err != nil {
			log.Fatalf("Error setting proxy-arp in VPP, %s", err.Error())
		}

	}
}

func (c *Client) configDHCPRelay() {
	net, err := netip.ParsePrefix(c.config.TapNetworkPrefix)
	if err != nil {
		log.Fatalf("Error parsing Tap IPv4Pool, %s", err.Error())
	}

	first := net.Addr().Next()
	second := first.Next()

	vppIPFirst := &ip_types.Address{
		Af: ip_types.ADDRESS_IP4,
		Un: ip_types.AddressUnionIP4(ip_types.IP4Address{
			first.As4()[0], first.As4()[1], first.As4()[2], first.As4()[3],
		}),
	}

	vppIPSecond := &ip_types.Address{
		Af: ip_types.ADDRESS_IP4,
		Un: ip_types.AddressUnionIP4(ip_types.IP4Address{
			second.As4()[0], second.As4()[1], second.As4()[2], second.As4()[3],
		}),
	}

	// Create Tap interface
	var swIf int
	swIf, err = c.createTapInterface(vppIPSecond, uint8(net.Bits()))
	if err != nil {
		log.Fatalf("Error creating Tap interface, %s", err.Error())
	}

	// Set Tap interface up
	err = c.setInterfaceUp(swIf)
	if err != nil {
		log.Fatalf("Error setting Tap interface up, %s", err.Error())
	}

	// Add first IPv4 from net to early created Tap
	err = c.setInterfaceAddrIPv4(swIf, vppIPFirst, uint8(net.Bits()))
	if err != nil {
		log.Fatalf("Error adding IPv4 to tap interface, %s", err.Error())
	}

	// Enable DHCP Proxy to External Server
	err = c.setProxyDHCPv4(vppIPFirst, vppIPSecond)
	if err != nil {
		log.Fatalf("Error setting up DHCPv4 proxy, %s", err.Error())
	}
}

func (c *Client) configCPEInterfaces() {
	for _, v := range c.ifaces {
		// UP State to iface
		err := c.setInterfaceUp(v.VPPSrcIface)
		if err != nil {
			log.Fatalf("Error setting up interface, %s", err.Error())
		}

		// Test if it's a sub-interface
		swIf := v.VPPSrcIface
		if v.IsSubIf {
			if v.HasQinQ {
				// Add QinQ VLAN
				swIf, err = c.createQinQInterface(v.VPPSrcIface, (v.OuterVLAN<<12)+v.InnerVLAN,
					v.OuterVLAN, v.InnerVLAN)

				if err != nil {
					log.Fatalf("Error creating sub-interface QinQ, %s", err.Error())
				}
			} else {
				// Modify ID using an autogenerated
				swIf, err = c.createVlanInterface(v.VPPSrcIface, v.OuterVLAN, v.OuterVLAN)
				if err != nil {
					log.Fatalf("Error creating sub-interface VLAN, %s", err.Error())
				}
			}

			// UP State to sub-interface
			err := c.setInterfaceUp(swIf)
			if err != nil {
				log.Fatalf("Error setting up interface, %s", err.Error())
			}
		}
		// Set MTU
		err = c.setInterfaceMTU(swIf, v.MTU)
		if err != nil {
			log.Fatalf("Error setting MTU in interface, %s", err.Error())
		}

		// Set Unnumbered to loopback
		err = c.setInterfaceUnnumbered(swIf, c.gwLoopSwIf)
		if err != nil {
			log.Fatalf("Error setting unnumbered interface, %s", err.Error())
		}

		if c.config.EnableProxyARP {
			// Enable ProxyARP in interface
			c.setInterfaceProxyARP(swIf, true)
		}
	}
}

func (c *Client) configIPv4GwLoopback() {
	// Create loopback iface
	var err error
	c.gwLoopSwIf, err = c.createLoopackIface()
	if err != nil {
		log.Fatalf("Error creating loopback interface, %s", err.Error())
	}
	// Set loopback iface up
	err = c.setInterfaceUp(c.gwLoopSwIf)
	if err != nil {
		log.Fatalf("Error setting up loopback interface, %s", err.Error())
	}
	// Iterate over Gw IPv4 and set it to created loopback
	for _, v := range c.config.GatewayIfaceAddrs {
		ipv4 := net.ParseIP(v)
		if ipv4 == nil {
			log.Fatalf("Gateway IPv4 %s is not possible to parse", v)
		}

		vppip := ip_types.Address{
			Af: ip_types.ADDRESS_IP4,
			Un: ip_types.AddressUnionIP4(ip_types.IP4Address{
				ipv4[12], ipv4[13], ipv4[14], ipv4[15],
			}),
		}

		// Set IPv4 to loopback
		err = c.setInterfaceAddrIPv4(c.gwLoopSwIf, &vppip, 32)
		if err != nil {
			log.Fatalf("Error setting IPv4 in loopback interface, %s", err.Error())
		}
	}
}
