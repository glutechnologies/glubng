package core

import "github.com/glutechnologies/glubng/pkg/vpp"

type Sessions struct {
	sessions map[string]*Session
	vpp      vpp.Client
}

type Session struct {
	Iface int // VPP Iface
	IPv4  string
}

func (s *Sessions) Init(vpp *vpp.Client) {
	// Init vpp client
	s.vpp = *vpp
}

func (s *Sessions) AddSession(ses *Session) {
	s.sessions[ses.IPv4] = ses
	s.vpp.AddSession()
}

func (s *Sessions) RemoveSession(ipv4 string) {
	delete(s.sessions, ipv4)
	s.vpp.RemoveSession()
}

func (s *Sessions) GetSession(ipv4 string) *Session {
	return s.sessions[ipv4]
}
