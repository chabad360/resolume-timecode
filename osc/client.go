package osc

import (
	"fmt"
	"net"
)

// Client enables you to send OSC packets. It sends OSC messages and bundles to
// the given IP address and port.
type Client struct {
	ip    string
	port  int
	laddr *net.UDPAddr
}

// NewClient creates a new OSC client. The Client is used to send OSC
// messages and OSC bundles over an UDP network connection. The `ip` argument
// specifies the IP address and `port` defines the target port where the
// messages and bundles will be send to.
func NewClient(ip string, port int) *Client {
	return &Client{ip: ip, port: port, laddr: nil}
}

// IP returns the IP address.
func (c *Client) IP() string { return c.ip }

// SetIP sets a new IP address.
func (c *Client) SetIP(ip string) { c.ip = ip }

// Port returns the port.
func (c *Client) Port() int { return c.port }

// SetPort sets a new port.
func (c *Client) SetPort(port int) { c.port = port }

// SetLocalAddr sets the local address.
func (c *Client) SetLocalAddr(ip string, port int) error {
	laddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return err
	}
	c.laddr = laddr
	return nil
}

// Send sends an OSC Bundle or an OSC Message.
func (c *Client) Send(packet Packet) error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", c.ip, c.port))
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp", c.laddr, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	data, err := packet.MarshalBinary()
	if err != nil {
		return err
	}

	if _, err = conn.Write(data); err != nil {
		return err
	}
	return nil
}
