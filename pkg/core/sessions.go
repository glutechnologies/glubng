package core

import (
	"log"
	"net"

	"github.com/glutechnologies/glubng/pkg/vpp"
)

type Sessions struct {
	sessions map[string]*Session
	vpp      vpp.Client
}

type Session struct {
	Iface int // VPP Iface
	IPv4  net.IP
}

func (s *Sessions) Init(vpp *vpp.Client) {
	// Init vpp client
	s.vpp = *vpp
	s.sessions = make(map[string]*Session)
}

func (s *Sessions) AddSession(ses *Session) {
	s.sessions[ses.IPv4.String()] = ses
	s.vpp.AddSession(ses.IPv4, uint32(ses.Iface))
}

func (s *Sessions) RemoveSession(ipv4 string) {
	ses := s.sessions[ipv4]
	if ses == nil {
		log.Printf("Session with IPv4 %s not exists", ipv4)
		return
	}
	s.vpp.RemoveSession(ses.IPv4, uint32(ses.Iface))
	delete(s.sessions, ipv4)
}

func (s *Sessions) GetSession(ipv4 string) *Session {
	return s.sessions[ipv4]
}
