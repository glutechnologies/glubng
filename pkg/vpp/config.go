package vpp

import (
	"bytes"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

// VPP related configuration
type VPPConfig struct {
	SrcVPPSocket      string
	UplinkIfaceName   string
	UplinkIfaceIPv4   string
	GatewayIfaceAddrs []string
	EnableProxyARP    bool
}

// VPP CPE Interfaces
type Iface struct {
	VPPSrcIface int
	IsSubIf     bool
	HasQinQ     bool
	OuterVLAN   int
	InnerVLAN   int
}

func (c *Client) LoadIfacesConfig() {
	body, err := os.ReadFile(c.ifacesFile)

	if err != nil {
		log.Fatalf("Error loading configuration file")
	}

	_, err = toml.Decode(string(body), &c.ifaces)

	if err != nil {
		log.Fatalf("Error decoding configuration file")
	}
}

func (c *Client) WriteIfacesConfig() {

	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(c.ifaces)

	if err != nil {
		log.Fatalf("Error encoding configuration in TOML")
	}

	err = os.WriteFile(c.ifacesFile, buf.Bytes(), 0666)

	if err != nil {
		log.Fatalf("Error writing configuration file")
	}
}