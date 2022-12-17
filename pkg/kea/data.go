package kea

import (
	"encoding/json"
	"log"
	"net"

	"github.com/glutechnologies/glubng/pkg/utils"
)

const CALLOUT_LEASE4_SELECT = 1
const CALLOUT_LEASE4_RENEW = 2
const CALLOUT_LEASE4_RELEASE = 3
const CALLOUT_LEASE4_DECLINE = 4
const CALLOUT_LEASE4_EXPIRE = 5
const CALLOUT_LEASE4_RECOVER = 6
const CALLOUT_PKT4_CIRCUIT_ID = 7

type Envelope struct {
	Callout int             `json:"callout"`
	Lease   json.RawMessage `json:"lease"`
	Subnet  json.RawMessage `json:"subnet"`
	Query   json.RawMessage `json:"query"`
}

type Lease struct {
	State     string `json:"state"`
	IsExpired bool   `json:"is-expired"`
	Address   string `json:"address"`
	Hostname  string `json:"hostname"`
	Cltt      int    `json:"cltt"`
	ValidLft  int    `json:"valid-lft"`
}

type Query struct {
	Type         string `json:"type"`
	Interface    string `json:"interface"`
	IfIndex      int    `json:"if-index"`
	HwAddr       string `json:"hw-addr"`
	HwAddrType   string `json:"hw-addr-type"`
	HwAddrSource string `json:"hw-addr-source"`
	Option60     string `json:"option60"`
	Option82     string `json:"option82"`
	Option82CID  string `json:"option82-circuit-id"`
	Option82RID  string `json:"option82-remote-id"`
}

type Subnet struct {
	Name   string `json:"name"`
	Prefix string `json:"prefix"`
	Len    int    `json:"len"`
}

type KeaResult struct {
	Callout int
	Query   Query
	Subnet  Subnet
	Lease   Lease
}

type KeaResponse struct {
	FlexId string `json:"flex-id"`
}

func sendResponse(k *KeaSocket, r *KeaResult, conn net.Conn) {
	// Prepare Kea Result

	if len(r.Query.Option82CID) > 0 {
		ifSw, err := utils.ConvertCIDToInt(r.Query.Option82CID)

		if err != nil {
			log.Printf("Error parsing response Kea, %s", err.Error())
			return
		}

		resp := &KeaResponse{FlexId: k.ifacesSwIf[int(ifSw)].FlexId}

		e := json.NewEncoder(conn)
		err = e.Encode(resp)

		if err != nil {
			log.Printf("Error sending response Kea, %s", err.Error())
			return
		}
	}
}

func processDataFromConnection(k *KeaSocket, env *Envelope) *KeaResult {
	var r KeaResult
	r.Callout = env.Callout
	switch env.Callout {
	case CALLOUT_LEASE4_RENEW, CALLOUT_LEASE4_SELECT:
		if err := json.Unmarshal(env.Lease, &r.Lease); err != nil {
			log.Println(err)
		}
		if err := json.Unmarshal(env.Subnet, &r.Subnet); err != nil {
			log.Println(err)
		}
		if err := json.Unmarshal(env.Query, &r.Query); err != nil {
			log.Println(err)
		}
		// Send message to other goroutines
		k.Message <- r
	case CALLOUT_LEASE4_RELEASE, CALLOUT_LEASE4_DECLINE:
		if err := json.Unmarshal(env.Lease, &r.Lease); err != nil {
			log.Println(err)
		}
		if err := json.Unmarshal(env.Query, &r.Query); err != nil {
			log.Println(err)
		}
		// Send message to other goroutines
		k.Message <- r
	case CALLOUT_LEASE4_EXPIRE, CALLOUT_LEASE4_RECOVER:
		if err := json.Unmarshal(env.Lease, &r.Lease); err != nil {
			log.Println(err)
		}
		// Send message to other goroutines
		k.Message <- r
	case CALLOUT_PKT4_CIRCUIT_ID:
		if err := json.Unmarshal(env.Query, &r.Query); err != nil {
			log.Println(err)
		}
	default:
		log.Printf("Process from connection, unknown message type: %q", env.Callout)
	}

	return &r
}
