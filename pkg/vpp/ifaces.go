package vpp

import (
	"go.fd.io/govpp/binapi/arp"
	interfaces "go.fd.io/govpp/binapi/interface"
	"go.fd.io/govpp/binapi/interface_types"
	"go.fd.io/govpp/binapi/ip_types"
	"go.fd.io/govpp/binapi/tapv2"
)

func (c *Client) setInterfaceUp(swIf int) error {
	// Set interface up
	req := &interfaces.SwInterfaceSetFlags{SwIfIndex: interface_types.InterfaceIndex(swIf),
		Flags: interface_types.IF_STATUS_API_FLAG_ADMIN_UP,
	}

	reply := &interfaces.SwInterfaceSetFlagsReply{}

	if err := c.ch.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func (c *Client) setInterfaceMTU(swIf int, mtu uint32) error {
	// Set interface MTU
	vppmtu := []uint32{mtu, 0, 0, 0}
	req := &interfaces.SwInterfaceSetMtu{
		SwIfIndex: interface_types.InterfaceIndex(swIf),
		Mtu:       vppmtu,
	}

	reply := &interfaces.SwInterfaceSetMtuReply{}

	if err := c.ch.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func (c *Client) setInterfaceProxyARP(swIf int, enable bool) error {
	// Set interface proxy
	req := &arp.ProxyArpIntfcEnableDisable{
		SwIfIndex: interface_types.InterfaceIndex(swIf),
		Enable:    enable,
	}

	reply := &arp.ProxyArpIntfcEnableDisableReply{}

	if err := c.ch.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func (c *Client) setInterfaceAddrIPv4(swIf int, ipv4 *ip_types.Address, len uint8) error {
	req := &interfaces.SwInterfaceAddDelAddress{
		SwIfIndex: interface_types.InterfaceIndex(swIf),
		IsAdd:     true,
		Prefix: ip_types.AddressWithPrefix{
			Address: *ipv4,
			Len:     len,
		},
	}
	reply := &interfaces.SwInterfaceAddDelAddressReply{}

	if err := c.ch.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func (c *Client) setInterfaceUnnumbered(swIf int, toSwIf int) error {
	req := &interfaces.SwInterfaceSetUnnumbered{
		SwIfIndex:           interface_types.InterfaceIndex(toSwIf),
		UnnumberedSwIfIndex: interface_types.InterfaceIndex(swIf),
		IsAdd:               true,
	}

	reply := &interfaces.SwInterfaceSetUnnumberedReply{}

	if err := c.ch.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func (c *Client) createLoopackIface() (int, error) {
	req := &interfaces.CreateLoopback{}
	reply := &interfaces.CreateLoopbackReply{}

	if err := c.ch.SendRequest(req).ReceiveReply(reply); err != nil {
		return 0, err
	}

	return int(reply.SwIfIndex), nil
}

func (c *Client) createTapInterface(ipv4 *ip_types.Address, len uint8) (int, error) {
	req := &tapv2.TapCreateV2{
		ID:               0,
		UseRandomMac:     true,
		HostIfNameSet:    true,
		HostIfName:       c.config.TapIfaceName,
		HostIP4PrefixSet: true,
		HostIP4Prefix:    ip_types.IP4AddressWithPrefix{Address: ipv4.Un.GetIP4(), Len: len},
	}
	reply := &tapv2.TapCreateV2Reply{}

	if err := c.ch.SendRequest(req).ReceiveReply(reply); err != nil {
		return 0, err
	}

	return int(reply.SwIfIndex), nil
}

func (c *Client) createVlanInterface(swIf int, id int, vlan int) (int, error) {
	req := &interfaces.CreateSubif{
		SwIfIndex:   interface_types.InterfaceIndex(swIf),
		SubID:       uint32(id),
		OuterVlanID: uint16(vlan),
		SubIfFlags:  interface_types.SUB_IF_API_FLAG_ONE_TAG | interface_types.SUB_IF_API_FLAG_EXACT_MATCH,
	}

	reply := &interfaces.CreateSubifReply{}

	if err := c.ch.SendRequest(req).ReceiveReply(reply); err != nil {
		return 0, err
	}

	return int(reply.SwIfIndex), nil
}

func (c *Client) createQinQInterface(swIf int, id int, outerVlan int, innerVlan int) (int, error) {
	req := &interfaces.CreateSubif{
		SwIfIndex:   interface_types.InterfaceIndex(swIf),
		SubID:       uint32(id),
		OuterVlanID: uint16(outerVlan),
		InnerVlanID: uint16(innerVlan),
		SubIfFlags:  interface_types.SUB_IF_API_FLAG_TWO_TAGS | interface_types.SUB_IF_API_FLAG_EXACT_MATCH,
	}

	reply := &interfaces.CreateSubifReply{}

	if err := c.ch.SendRequest(req).ReceiveReply(reply); err != nil {
		return 0, err
	}

	return int(reply.SwIfIndex), nil
}
