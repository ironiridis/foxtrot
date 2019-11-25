package cip

import "net"

// Accept completes an inbound CIP connection.
func Accept(l net.Listener) (*Conn, error) {
	j := NewJoins()
	return AcceptWithJoins(l, j)
}

// AcceptWithJoins completes an inbound CIP connection. `j` must be an existing
// Joins structure which will be synchronized on connect.
func AcceptWithJoins(l net.Listener, j *Joins) (*Conn, error) {
	var err error
	c := new(Conn)

	s, err := l.Accept()
	if err != nil {
		return nil, err
	}
	tcpsock, ok := s.(*net.TCPConn)
	if !ok {
		s.Close()
		return nil, ErrorInvalidNetwork
	}
	c.socket = tcpsock
	c.Joins = j
	go c.readLoop()
	c.sendHello()
	return c, nil
}
