package vpp

import (
	"go.fd.io/govpp/binapi/dhcp"
	"go.fd.io/govpp/binapi/ip_types"
)

func (c *Client) setProxyDHCPv4(src *ip_types.Address, dst *ip_types.Address) error {
	req := &dhcp.DHCPProxyConfig{
		IsAdd:          true,
		DHCPServer:     *dst,
		DHCPSrcAddress: *src,
	}

	reply := &dhcp.DHCPProxyConfigReply{}

	if err := c.ch.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}
