package kea

import (
	"encoding/json"
	"log"
)

const CALLOUT_LEASE4_SELECT = 1
const CALLOUT_LEASE4_RENEW = 2
const CALLOUT_LEASE4_RELEASE = 3
const CALLOUT_LEASE4_DECLINE = 4
const CALLOUT_LEASE4_EXPIRE = 5
const CALLOUT_LEASE4_RECOVER = 6

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

func processDataFromConnection(k *KeaSocket, env *Envelope) {
	var r KeaResult
	r.Callout = env.Callout
	switch env.Callout {
	case CALLOUT_LEASE4_RENEW:
	case CALLOUT_LEASE4_RELEASE:
	case CALLOUT_LEASE4_SELECT:
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
	case CALLOUT_LEASE4_DECLINE:
		if err := json.Unmarshal(env.Lease, &r.Lease); err != nil {
			log.Println(err)
		}
		if err := json.Unmarshal(env.Query, &r.Query); err != nil {
			log.Println(err)
		}
		// Send message to other goroutines
		k.Message <- r

	case CALLOUT_LEASE4_EXPIRE:
	case CALLOUT_LEASE4_RECOVER:
		if err := json.Unmarshal(env.Lease, &r.Lease); err != nil {
			log.Println(err)
		}
		// Send message to other goroutines
		k.Message <- r
	default:
		log.Printf("Process from connection, unknown message type: %q", env.Callout)
	}
}
